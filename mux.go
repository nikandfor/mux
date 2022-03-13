package mux

import (
	"net/http"
	pathpkg "path"
)

type (
	Mux struct {
		//	root *page
		meth map[string]*page
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

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := getContext()
	defer putContext(c)

	c.Context = req.Context()
	c.ResponseWriter = w
	c.Request = req

	h := m.match(req.Method, req.RequestURI, c)
	if h == nil {
		// TODO: 404
		return
	}

	err := h(c)
	_ = err
}

func (m *Mux) handle(meth, path string, h HandlerFunc) {
	if m.meth == nil {
		m.meth = make(map[string]*page)
	}

	root, ok := m.meth[meth]

	page := m.put(path, 0, root)
	if !ok {
		m.meth[meth] = page
	}

	if page.h != nil {
		panic("routing collision")
	}

	page.h = h
}

func (m *Mux) match(meth, path string, c *Context) HandlerFunc {
	root, ok := m.meth[meth]
	if !ok {
		return nil
	}

	page := m.get(path, 0, root)
	if page == nil {
		return nil
	}

	return page.h
}
