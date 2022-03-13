package mux

import (
	"net/http"
	pathpkg "path"
)

type (
	Mux struct {
		meth map[string]*page

		buf [256]page
		i   int
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
	m.put(meth, path, h)
}

func (m *Mux) match(meth, path string, c *Context) HandlerFunc {
	return m.get(meth, path, c)
}
