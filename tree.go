package mux

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type (
	node struct {
		pref string

		h      int32
		branch bool

		len int
		x   [pagesize]uint32 // link << 8 | char
	}
)

const pagesize = 4

var ErrRouting = errors.New("routing collision")

var spaces = "                                                                                                                              "

func (m *Mux) put(meth, path string, h int32) error {
	if m.meth == nil {
		m.meth = make(map[string]uint32)
	}

	i, ok := m.meth[meth]
	if !ok {
		i = m.new(path)
		m.meth[meth] = i
	}

	return m.nodePut(path, 0, i, h)
}

func (m *Mux) nodePut(path string, st int, i uint32, h int32) error {
	n := &m.nodes[i]

	/*
		x := '_'
		if n.branch {
			x = 'X'
		}

		defer func(st int) {
			if n.branch {
				x = 'X'
			}

			fmt.Printf("<< node %3x  put  %-24q  into %-10q  h%3x %c len %d  %v  from %v\n", i, path[st:], n.pref, n.h, x, n.len, ll(n.x, n.len), loc.Callers(1, 2))
		}(st)

		fmt.Printf(">> node %3x  put  %-24q  into %-10q  h%3x %c len %d  %v  from %v\n", i, path[st:], n.pref, n.h, x, n.len, ll(n.x, n.len), loc.Callers(1, 2))
	*/

	c := common(n.pref, path[st:])

	if c != len(n.pref) { // split trie
		// allocate new
		j := m.new(n.pref[c:])
		n = &m.nodes[i] // renew pointer since underlaying array may change while new
		moved := &m.nodes[j]

		// move current (to not change link to i'th node)
		moved.h = n.h
		moved.len = n.len
		moved.x = n.x

		n.pref = n.pref[:c]
		n.h = -1

		n.x = [pagesize]uint32{}

		n.len = 1
		n.x[0] = j<<8 | uint32(moved.pref[0])

		//	fmt.Printf("== node %x  put  %-10q  into %-10q  h%3x len %d  %v  from %v\n", i, path[st:], n.pref, n.h, n.len, ll(n.x, n.len), loc.Callers(1, 2))
	}

	st += c

	if st == len(path) {
		if n.h >= 0 {
			return ErrRouting
		}

		n.h = h

		return nil
	}

	j := search(n, path[st])

	if n.branch || j < n.len && path[st] == byte(n.x[j]) {
		if j == n.len || path[st] < byte(n.x[j]) && j != 0 {
			j--
		}

		return m.nodePut(path, st, n.x[j]>>8, h)
	}

	if n.len < pagesize {
		m.insert(path, st, i, h, j)
		return nil
	}

	// split page

	li := m.new("")
	ri := m.new("")

	l := &m.nodes[li]
	r := &m.nodes[ri]
	n = &m.nodes[i] // retake pointer after append

	mid := n.len / 2

	copy(l.x[:], n.x[:mid])
	copy(r.x[:n.len-mid], n.x[mid:])
	l.len = mid
	r.len = n.len - mid
	l.branch = n.branch
	r.branch = n.branch

	n.x[0] = uint32(li<<8) | l.x[0]&0xff
	n.x[1] = uint32(ri<<8) | r.x[0]&0xff
	n.len = 2

	n.branch = true

	var sub uint32
	if j <= mid {
		sub = li
	} else {
		sub = ri
		j -= mid
	}

	m.insert(path, st, sub, h, j)

	return nil
}

func (m *Mux) insert(path string, st int, i uint32, h int32, j int) {
	ch := m.new(path[st:])
	m.nodes[ch].h = h

	n := &m.nodes[i]

	copy(n.x[j+1:], n.x[j:n.len])
	n.len++
	n.x[j] = uint32(ch<<8) | uint32(path[st])
}

func (m *Mux) get(meth, path string) (h int32) {
	i, ok := m.meth[meth]
	if !ok {
		return -1
	}

	return m.nodeGet(path, 0, i)
}

func (m *Mux) nodeGet(path string, st int, i uint32) int32 {
	n := &m.nodes[i]

	var j int
	//	defer func() {
	//		fmt.Printf(">> node %3x  get  %-24q  j %x  %v\n", i, path[st:], j, ll(n.x, n.len))
	//	}()

	c := common(n.pref, path[st:])

	if c != len(n.pref) {
		return -1
	}

	st += c

	if st == len(path) {
		return n.h
	}

	if n.len == 0 {
		return -1
	}

	j = search(n, path[st])

	if j == n.len || path[st] < byte(n.x[j]) && j != 0 {
		j--
	}

	return m.nodeGet(path, st, n.x[j]>>8)
}

func (m *Mux) new(pref string) (i uint32) {
	i = uint32(len(m.nodes))
	//	fmt.Printf("++ new  %3x  pref %-24q  from %v\n", i, pref, loc.Callers(1, 2))
	m.nodes = append(m.nodes, node{
		pref: pref,
		h:    -1,
	})
	return i
}

func (m *Mux) dumpString(i uint32) string {
	var b bytes.Buffer

	m.dump(&b, i, 0, 0)

	//	fmt.Fprintf(&b, "======\n")

	//	for i, n := range m.nodes {
	//		fmt.Fprintf(&b, "node %4x  pref %-24q  func %3x  childs %d  %v\n", i, n.pref, n.h, n.len, ll(n.x, n.len))
	//	}

	return b.String()
}

func (m *Mux) dump(w io.Writer, i uint32, c byte, d int) {
	if d > 8 {
		return
	}

	n := &m.nodes[i]

	x := '_'
	if n.branch {
		x = 'X'
	}

	fmt.Fprintf(w, "%vnode%v %4x  pref %-24q  func %4x  %c  childs %d  %v\n", spaces[:2*d], spaces[:2*(10-d)], i, n.pref, n.h, x, n.len, ll(n.x, n.len))
	for j := 0; j < n.len; j++ {
		m.dump(w, n.x[j]>>8, byte(n.x[j]), d+1)
	}
}

func search(n *node, q byte) (j int) {
	l, r := 0, n.len

	for l < r {
		j = int(uint(l+r) >> 1)

		if q <= byte(n.x[j]) {
			r = j
		} else {
			l = j + 1
		}
	}

	return l
}

func common(a, b string) (c int) {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}

	for c < m && a[c] == b[c] {
		c++
	}

	return c
}

func ll(x [pagesize]uint32, l int) string {
	var b bytes.Buffer

	b.WriteByte('[')

	for i := range x[:l] {
		fmt.Fprintf(&b, " %q:%x", byte(x[i]), x[i]>>8)
	}

	b.WriteByte(']')

	return b.String()
}
