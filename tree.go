package mux

import (
	"bytes"
	"fmt"
	"io"
	"sort"
)

const (
	pagesize = 16

	none int32 = -1
)

var spaces = "                                                                                                                              "

type (
	page struct {
		pref string

		h HandlerFunc

		k []byte
		s []*page
	}
)

func (m *Mux) putHandler(path string, st int, p *page, h HandlerFunc) {
	p = m.put(path, st, p)
	if p.h != nil {
		panic("already set")
	}

	p.h = h

	//	if m.root == nil {
	//		m.root = p
	//	}
}

func (m *Mux) getHandler(path string, st int, p *page) HandlerFunc {
	if p == nil {
		return nil
	}

	p = m.get(path, p)
	if p == nil {
		return nil
	}

	return p.h
}

func (m *Mux) get(path string, p *page) *page {
	if p == nil {
		return nil
	}

loop:
	for {
		c := common(p.pref, path)

		if c != len(p.pref) {
			return nil
		}

		if c == len(path) {
			return p
		}

		for j, k := range p.k {
			if k == path[c] {
				p = p.s[j]
				path = path[c:]
				continue loop
			}
		}

		return nil

		//	j := p.k[path[c]-0x20]
		//	if j == -1 {
		//		return nil
		//	}

		//	p = p.s[j]
		//	path = path[c:]

		//	st = c
	}
}

func (m *Mux) put(path string, st int, p *page) (sub *page) {
	if p == nil {
		return m.new(path[st:])
	}

	for {
		c := common(p.pref, path[st:])

		if c != len(p.pref) {
			sub = m.new(p.pref[c:])
			sub.h = p.h
			sub.k = p.k
			sub.s = p.s

			p.pref = p.pref[:c]
			p.h = nil
			p.k = nil
			p.s = nil

			p.setsub(sub.pref[0], sub)
		}

		c += st

		if c == len(path) {
			return p
		}

		sub = p.sub(path[c])
		if sub != nil {
			st = c
			p = sub

			continue
		}

		sub = m.new(path[c:])

		p.setsub(sub.pref[0], sub)

		return sub
	}
}

func common(a, b string) (c int) {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}

	_ = a[m-1] == b[m-1]

	for c < m && a[c] == b[c] {
		c++
	}

	return c
}

func (m *Mux) dumpString(p *page) string {
	if p == nil {
		return "<nil>"
	}

	var b bytes.Buffer

	m.dump(&b, p, 0, 0)

	return b.String()
}

func (m *Mux) dump(w io.Writer, p *page, d, st int) {
	const maxdepth = 10
	if d > maxdepth {
		fmt.Fprintf(w, "%v...\n", spaces[:2*d])
		return
	}

	valPad := 12 - st
	if valPad < 0 {
		valPad = 0
	}

	fmt.Fprintf(w, "   dump %v%d%v  pref %v%-24q%v  h %p  %v\n", spaces[:2*d], d, spaces[:2*(maxdepth-d)], spaces[:st], p.pref, spaces[:valPad], p.h, m.ll(p))
	//	fmt.Fprintf(w, "%vpage%v %4x  pref %-24q  val %4x  %c  childs %d  %v\n", spaces[:2*d], spaces[:2*(10-d)], i, m.p[i].pref, m.p[i].val, x, m.p[i].len, m.ll(i))
	for _, sub := range p.s {
		m.dump(w, sub, d+1, st+len(p.pref))
	}
}

func (m *Mux) ll(p *page) string {
	var b bytes.Buffer

	b.WriteByte('[')

	for j, k := range p.k {
		fmt.Fprintf(&b, " %q:%d", byte(k), count(p.s[j]))
	}

	b.WriteByte(']')

	return b.String()
}

func (m *Mux) new(pref string) (p *page) {
	if m.i < len(m.buf) {
		p = &m.buf[m.i]
		m.i++

		*p = page{
			pref: pref,
			//	k:    nonek,
		}

		return p
	}

	p = &page{
		pref: pref,
		//	k:    nonek,
	}

	return p
}

func (p *page) sub(f byte) *page {

	if !sort.IsSorted(bysize{p}) {
		panic("not sorter")
	}

	for j, k := range p.k {
		if k == f {
			return p.s[j]
		}
	}

	return nil
}

func (p *page) setsub(f byte, sub *page) *page {
	p.k = append(p.k, f)
	p.s = append(p.s, sub)

	sort.Sort(bysize{p})

	return nil
}

func (p *page) sub1(path string) *page {
	j := p.k[path[0]-0x20]
	//	if j == -1 {
	//		return nil
	//	}

	return p.s[j]
}

func (p *page) setSub1(path string, sub *page) {
	//	p.k[path[0]-0x20] = int8(len(p.s))
	//	p.s = append(p.s, sub)
}

func (p *page) sub0(path string) *page {
	for j, k := range p.k {
		//	if k == int8(path[0]) {
		_ = k
		return p.s[j]
		//	}
	}

	return nil
}

func (p *page) setSub0(path string, sub *page) {
	//	p.k = append(p.k, path[0])
	//	p.k[len(p.s)] = int8(path[0])
	p.s = append(p.s, sub)

	sort.Sort(bysize{p})
}

func (p *page) Len() int           { return len(p.s) }
func (p *page) Less(i, j int) bool { return p.k[i] < p.k[j] }
func (p *page) Swap(i, j int) {
	p.s[i], p.s[j] = p.s[j], p.s[i]
	p.k[i], p.k[j] = p.k[j], p.k[i]
}

type bysize struct {
	*page
}

func (p bysize) Len() int           { return len(p.s) }
func (p bysize) Less(i, j int) bool { return count(p.s[i]) > count(p.s[j]) }
func (p bysize) Swap(i, j int) {
	p.s[i], p.s[j] = p.s[j], p.s[i]
	p.k[i], p.k[j] = p.k[j], p.k[i]
}

func count(p *page) (sum int) {
	if p.h != nil {
		sum++
	}

	for _, s := range p.s {
		sum += count(s)
	}

	return
}
