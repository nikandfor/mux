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

		h HandlerFunc
		//	p *page

		x [0x80 - 0x20]*page
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

	p = m.get(path, st, p)
	if p == nil {
		return nil
	}

	return p.h
}

func (m *Mux) get(path string, st int, p *page) *page {
	if p == nil {
		return nil
	}

	for {
		c := common(p.pref, path[st:])

		if c != len(p.pref) {
			return nil
		}

		c += st

		if c == len(path) {
			return p
		}

		sub := p.x[path[c]-0x20]
		if sub == nil {
			return nil
		}

		st = c
		p = sub
	}
}

func (m *Mux) put(path string, st int, p *page) (sub *page) {
	if p == nil {
		return &page{
			pref: path[st:],
		}
	}

	for {
		c := common(p.pref, path[st:])

		if c != len(p.pref) {
			sub = &page{
				pref: p.pref[c:],
				x:    p.x,
				h:    p.h,
				//	p:    p.p,
			}

			p.pref = p.pref[:c]
			p.h = nil
			//	p.p = nil
			p.x = [0x80 - 0x20]*page{}
			p.x[sub.pref[0]-0x20] = sub
		}

		c += st

		if c == len(path) {
			return p
		}

		sub = p.x[path[c]-0x20]
		if sub != nil {
			st = c
			p = sub

			continue
		}

		sub = &page{
			pref: path[c:],
		}

		p.x[path[c]-0x20] = sub

		return sub
	}
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
	for _, sub := range p.x {
		if sub == nil {
			continue
		}

		m.dump(w, sub, d+1, st+len(p.pref))
	}
}

func (m *Mux) ll(p *page) string {
	var b bytes.Buffer

	b.WriteByte('[')

	for j, sub := range p.x {
		if sub == nil {
			continue
		}

		fmt.Fprintf(&b, " %q", byte(0x20+j))
	}

	b.WriteByte(']')

	return b.String()
}
