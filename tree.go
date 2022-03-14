package mux

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
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

		wildcard *page

		// wildcard page

		name    string
		re      *regexp.Regexp
		pattern string
	}
)

func (m *Mux) get(p *page, path string, c *Context) (leaf *page) {
	if p == nil {
		return nil
	}

	type wild struct {
		p      *page
		path   string
		params int
	}

	var wwbuf [20]wild
	ww := wwbuf[:0]

	for {
		if p == nil || len(p.pref) > len(path) || p.pref != path[:len(p.pref)] {
			if len(ww) == 0 {
				return nil
			}

			l := len(ww) - 1
			w := ww[l]
			ww = ww[:l]

			p, path = w.p, w.path

			//	println("pattern match", path, p.pattern, p.re)

			var val string
			if p.re != nil {
				if !p.re.MatchString(path) {
					p = nil
					continue
				}

				val = p.re.FindString(path)
			} else if p.pattern == "*" {
				val = path
			} else if p.pattern == "\\w+" {
				val = path[:index(path, 0, '/')]
			}

			if c != nil {
				c.Params = c.Params[:w.params]
				c.Params = append(c.Params, Param{
					Name:  p.name,
					Value: val,
				})
			}

			path = path[len(val):]

			continue
		} else if len(p.pref) == len(path) {
			return p
		}

		path = path[len(p.pref):]

		var sub *page

		for j, k := range p.k {
			if k == path[0] {
				sub = p.s[j]

				break
			}
		}

		if p.wildcard != nil {
			params := 0
			if c != nil {
				params = len(c.Params)
			}

			ww = append(ww, wild{
				p:      p.wildcard,
				path:   path,
				params: params,
			})
		}

		p = sub
	}
}

func (m *Mux) put(p *page, path string) (sub *page) {
	if p.pref == "" && p.name == "" { // new root node
		p.pref = path

		return p
	}

	for {
		c := common(p.pref, path)

		if c != len(p.pref) {
			cp := *p
			sub = &cp

			sub.pref = p.pref[c:]

			*p = page{
				pref: p.pref[:c],
			}

			p.setsub(sub.pref[0], sub)
		}

		if c == len(path) {
			return p
		}

		sub = p.sub(path[c])
		if sub != nil {
			path = path[c:]
			p = sub

			continue
		}

		sub = &page{
			pref: path[c:],
		}

		p.setsub(sub.pref[0], sub)

		return sub
	}
}

func common(a, b string) (c int) {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}

	_, _ = a[:m], b[:m]

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

	pref := p.pref
	if p.name != "" {
		pref = "{" + p.name + ":" + p.pattern + "}" + p.pref
	}

	valPad := 36 - st - len(pref)
	if valPad < 0 {
		valPad = 0
	}

	fmt.Fprintf(w, "   dump %v%d%v  pref %v%q%v  h %p  %v\n", spaces[:2*d], d, spaces[:2*(maxdepth-d)], spaces[:st], pref, spaces[:valPad], p.h, p.ll())
	//	fmt.Fprintf(w, "%vpage%v %4x  pref %-24q  val %4x  %c  childs %d  %v\n", spaces[:2*d], spaces[:2*(10-d)], i, m.p[i].pref, m.p[i].val, x, m.p[i].len, m.ll(i))
	for _, sub := range p.s {
		m.dump(w, sub, d+1, st+len(pref))
	}

	if p.wildcard != nil {
		m.dump(w, p.wildcard, d+1, st+len(pref))
	}
}

func (p *page) ll() string {
	var b bytes.Buffer

	b.WriteByte('[')

	for j, k := range p.k {
		fmt.Fprintf(&b, " %q:%d", byte(k), countHandlers(p.s[j]))
	}

	b.WriteByte(']')

	return b.String()
}

func (p *page) sub(f byte) *page {
	for j, k := range p.k {
		if k == f {
			return p.s[j]
		}
	}

	return nil
}

func (p *page) setsub(f byte, sub *page) {
	p.k = append(p.k, f)
	p.s = append(p.s, sub)
}

func countHandlers(p *page) (sum int) {
	if p.h != nil {
		sum++
	}

	for _, s := range p.s {
		sum += countHandlers(s)
	}

	if p.wildcard != nil {
		sum += countHandlers(p.wildcard)
	}

	return
}
