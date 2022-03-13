package mux

import (
	"bytes"
	"fmt"
	"io"
)

const (
	pagesize = 16

	none int32 = -1
)

var spaces = "                                                                                                                              "

type (
	page struct {
		pref string

		x [pagesize]int32

		len    int8
		branch bool

		val int32
	}
)

func (m *Mux) getVal(path string, root int32) int32 {
	p := m.get(path, 0, root)
	if p == none {
		return none
	}

	return m.p[p].val
}

func (m *Mux) putVal(path string, root, val int32) int32 {
	p := m.put(path, 0, root)
	m.p[p].val = val
	return p
}

func (m *Mux) get(path string, st int, root int32) (i int32) {
	if int(root) == len(m.p) {
		return none
	}

	i = root

	//	fmt.Printf(">> get  path %v%-*q  page%4x  ***\n", spaces[:st], 24-st, path[st:], i)
	//	defer func() {
	//		fmt.Printf("<< get  path %v%-*q  page%4x  *** from %v\n", spaces[:st], 24-st, path[st:], i, loc.Callers(1, 2))
	//	}()

	p := &m.p[i]

	for {

		c := common(p.pref, path[st:])

		//		fmt.Printf("   get  path %v%-*q  page%4x  cp%2x\n", spaces[:st], 24-st, path[st:], i, c)

		if c != len(p.pref) {
			return none
		}

		c += st

		if c == len(path) {
			return i
		}

		j := search(p, path[c])

		//		fmt.Printf("   get  path %v%-*q  page%4x  j %2x of %v  char %c\n", spaces[:st], 24-st, path[st:], i, j, m.ll(i), path[c])

		if p.branch || j < p.len && path[c] == byte(p.x[j]) {
			if p.branch && (j == p.len || j != 0 && path[c] < byte(p.x[j])) {
				j--
			}

			i = p.x[j] >> 8
			st = c

			p = &m.p[i]

			continue
		}

		return none
	}
}

func (m *Mux) put(path string, st int, root int32) (i int32) {
	if int(root) == len(m.p) {
		m.p = append(m.p, page{
			pref: path,
			val:  none,
		})

		return root
	}

	i = root

	for {
		c := common(m.p[i].pref, path[st:])

		if c != len(m.p[i].pref) {
			m.splitTrie(i, c)
		}

		c += st

		if c == len(path) {
			return i
		}

		j := search(&m.p[i], path[c])

		//		fmt.Printf("   put  path %v%-24q  page%4x  j %2x of %v  char %c\n", spaces[:st], path[st:], i, j, m.ll(i), path[c])

		if m.p[i].branch || j < m.p[i].len && path[c] == byte(m.p[i].x[j]) {
			if m.p[i].branch && (j == m.p[i].len || path[c] < byte(m.p[i].x[j]) && j != 0) {
				j--
			}

			i = m.p[i].x[j] >> 8
			st = c

			continue
		}

		if m.p[i].len < pagesize {
			return m.insert(i, path[c:], j)
		}

		i, j = m.splitPage(i, j)

		return m.insert(i, path[c:], j)
	}
}

func (m *Mux) splitPage(i int32, j int8) (int32, int8) {
	l := m.new()
	r := m.new()

	mid := m.p[i].len / 2

	copy(m.p[l].x[:], m.p[i].x[:mid])
	copy(m.p[r].x[:], m.p[i].x[mid:])
	m.p[l].len = mid
	m.p[r].len = m.p[i].len - mid

	m.p[l].branch = m.p[i].branch
	m.p[r].branch = m.p[i].branch

	m.p[i].x[0] = l<<8 | m.p[l].x[0]&0xff
	m.p[i].x[1] = r<<8 | m.p[r].x[0]&0xff
	m.p[i].len = 2

	m.p[i].branch = true

	if j <= mid {
		return l, j
	} else {
		return r, j - mid
	}
}

func (m *Mux) splitTrie(i int32, c int) {
	ch := m.new()

	m.p[ch] = m.p[i]
	m.p[ch].pref = m.p[i].pref[c:]

	m.p[i].pref = m.p[i].pref[:c]
	m.p[i].x[0] = ch<<8 | int32(m.p[ch].pref[0])
	m.p[i].len = 1
	m.p[i].branch = false

	m.p[i].val = none
}

func (m *Mux) insert(i int32, path string, j int8) (ch int32) {
	ch = m.new()
	m.p[ch].pref = path

	copy(m.p[i].x[j+1:], m.p[i].x[j:])
	m.p[i].x[j] = ch<<8 | int32(path[0])
	m.p[i].len++

	return ch
}

func (m *Mux) new() (i int32) {
	i = int32(len(m.p))
	m.p = append(m.p, page{
		val: none,
	})
	return i
}

func search(p *page, q byte) (j int8) {
	l, r := int8(0), p.len

	for l < r {
		j = l + (r-l)>>1

		if q <= byte(p.x[j]) {
			r = j
		} else {
			l = j + 1
		}
	}

	return l
}

func (m *Mux) search(i int32, q byte) (j int8) {
	l, r := int8(0), m.p[i].len

	for l < r {
		j = l + (r-l)>>1

		if q <= byte(m.p[i].x[j]) {
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

func (m *Mux) dumpString(i int32) string {
	if i == none {
		return "<none>"
	}
	if len(m.p) == 0 {
		return "<empty>"
	}

	var b bytes.Buffer

	m.dump(&b, i, 0, 0)

	//	fmt.Fprintf(&b, "======\n")

	//	for i, n := range m.p {
	//	}

	return b.String()
}

func (m *Mux) dump(w io.Writer, i int32, d, st int) {
	const maxdepth = 10
	if d > maxdepth {
		fmt.Fprintf(w, "%v...\n", spaces[:2*d])
		return
	}

	x := '_'
	if m.p[i].branch {
		x = 'X'
	}

	valPad := 12 - st
	if valPad < 0 {
		valPad = 0
	}

	fmt.Fprintf(w, "   dump %v%d%v  page%4x  pref %v%-24q  %vval%4x  %c  %v\n", spaces[:2*d], d, spaces[:2*(maxdepth-d)], i, spaces[:st], m.p[i].pref, spaces[:valPad], m.p[i].val, x, m.ll(i))
	//	fmt.Fprintf(w, "%vpage%v %4x  pref %-24q  val %4x  %c  childs %d  %v\n", spaces[:2*d], spaces[:2*(10-d)], i, m.p[i].pref, m.p[i].val, x, m.p[i].len, m.ll(i))
	for j := int8(0); j < m.p[i].len; j++ {
		m.dump(w, m.p[i].x[j]>>8, d+1, st+len(m.p[i].pref))
	}
}

func (m *Mux) ll(i int32) string {
	var b bytes.Buffer

	b.WriteByte('[')

	for j := range m.p[i].x[:m.p[i].len] {
		fmt.Fprintf(&b, " %q:%x", byte(m.p[i].x[j]), m.p[i].x[j]>>8)
	}

	b.WriteByte(']')

	return b.String()
}
