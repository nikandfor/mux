package mux

import (
	"fmt"
	"net/http"
	pathpkg "path"
	"regexp"
)

type (
	Mux struct {
		meth map[string]*page

		NotFoundHandler HandlerFunc
	}

	HandlerFunc func(c *Context) error
)

func New() *Mux {
	return &Mux{}
}

func (m *Mux) Handle(meth, path string, h HandlerFunc) {
	path = pathpkg.Clean(path)

	m.handle(meth, path, h)
}

func (m *Mux) Lookup(meth, path string, c *Context) HandlerFunc {
	return m.match(meth, path, c)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := GetContext(w, req)
	defer PutContext(c)

	h := m.match(req.Method, req.RequestURI, c)
	if h == nil {
		h = m.NotFoundHandler
	}

	_ = h(c)
}

func (m *Mux) handle(meth, path string, h HandlerFunc) {
	if m.meth == nil {
		m.meth = make(map[string]*page, 16)
	}

	p := m.meth[meth]
	if p == nil {
		p = &page{}
		m.meth[meth] = p
	}

	//	fmt.Printf("dump tree %v\n%v", meth, m.dumpString(p))
	//	fmt.Printf("add handler %v %v\n", meth, path)

	var i int
	for i < len(path) {
		st := i
		i = indexAny(path, i, "{:*")

		//	println("put", p.pref, "path_st_i_len", st, i, len(path), path[st:i])

		p = m.put(p, path[st:i])

		if i == len(path) {
			break
		}

		st = i + 1

		var name, pattern, re string
		switch path[i] {
		case '{':
			i = indexAny(path, i, "}:")
			name = path[st:i]
			if path[i] == ':' {
				st = i + 1
				i = index(path, i, '}')
				pattern = path[st:i]
				re = pattern
			} else {
				pattern = `\w+`
			}
			if path[i] != '}' {
				panic("no pattern closing bracket")
			}
			i++
		case ':':
			i = index(path, i, '/')
			name = path[st:i]
			pattern = `\w+`
		default:
			i = len(path)
			name = path[st:i]
			pattern = `*`
		}

		//	println("wname", name, "re", re, "path", path, "i", i)

		if p.wildcard != nil {
			if p.wildcard.name != name || p.wildcard.pattern != pattern {
				panic("wildcard collision")
			}
		} else {
			p.wildcard = &page{
				name:    name,
				pattern: pattern,
			}

			if re != "" {
				p.wildcard.re = regexp.MustCompile(`^` + re)
			}
		}

		p = p.wildcard
	}

	if p.h != nil {
		fmt.Printf("path collision: %v %v = %v ? %v\n", meth, path, h, p.h)
		fmt.Printf("dump tree %v\n%v", meth, m.dumpString(m.meth[meth]))

		panic("path collision: " + meth + " " + path)
	}

	p.h = h
}

func (m *Mux) match(meth, path string, c *Context) HandlerFunc {
	p := m.meth[meth]
	if p == nil {
		return nil
	}

	p = m.get(p, path, c)
	if p == nil {
		return nil
	}

	return p.h
}

func index(s string, st int, c byte) (i int) {
	for i = st; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}

	return i
}

func indexAny(s string, st int, c string) (i int) {
	for i = st; i < len(s); i++ {
		for _, c := range []byte(c) {
			if s[i] == c {
				return i
			}
		}
	}

	return i
}
