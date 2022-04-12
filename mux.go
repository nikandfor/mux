package mux

import (
	"net/http"
)

type (
	Mux struct {
		RouterGroup

		meth map[string]*node

		NotFound HandlerFunc

		ForwardedFor []string // list of headers to check for client address. used by Context.ClientIP
	}

	RouterGroup struct {
		m *Mux

		basePath string

		ms Middlewares
	}

	Middleware  func(next HandlerFunc) HandlerFunc
	Middlewares []Middleware

	HandlerFunc func(c *Context) error
)

func New() *Mux {
	m := &Mux{
		meth: make(map[string]*node),

		NotFound: SlashRedirector(NotFound),

		ForwardedFor: []string{"X-Forwarded-For", "X-Real-IP"},
	}

	m.RouterGroup = RouterGroup{
		m:        m,
		basePath: "/",
	}

	return m
}

func (g *RouterGroup) Use(ms ...Middleware) {
	g.ms.Append(ms...)
}

func (g *RouterGroup) Group(path string, ms ...Middleware) *RouterGroup {
	return &RouterGroup{
		m:        g.m,
		basePath: JoinPath(g.basePath, path),
		ms:       g.ms.Extend(ms...),
	}
}

func (g *RouterGroup) Handle(meth, path string, h HandlerFunc, ms ...Middleware) {
	ms = g.ms.Extend(ms...)

	g.m.handle(meth, JoinPath(g.basePath, path), ms, h)
}

func (m *Mux) Lookup(meth, path string, c *Context) (h HandlerFunc) {
	root := m.meth[meth]
	if root == nil {
		return nil
	}

	//	fmt.Printf("Lookup  %v %v\n", meth, path)

	node := m.get(root, path, c)
	if node == nil || node.h == nil {
		return nil
	}

	return node.h
}

func (m *Mux) handle(meth, path string, ms Middlewares, h HandlerFunc) {
	if path == "" || path[0] != '/' {
		panic("bad path")
	}

	root := m.meth[meth]
	if root == nil {
		root = &node{}
		m.meth[meth] = root
	}

	node := m.put(root, path)

	if node.h != nil {
		panic("path collision: " + path)
	}

	h = ms.Apply(h)

	node.h = h
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	defer FreeContext(c)

	c.m = m

	h := m.Lookup(req.Method, req.URL.Path, c)

	if h == nil {
		h = m.NotFound
		h = m.ms.Apply(h)
	}

	_ = h(c)
}

func (g *RouterGroup) GET(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodGet, path, h, ms...)
}

func (g *RouterGroup) HEAD(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodHead, path, h, ms...)
}

func (g *RouterGroup) POST(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodPost, path, h, ms...)
}

func (g *RouterGroup) PUT(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodPut, path, h, ms...)
}

func (g *RouterGroup) PATCH(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodPatch, path, h, ms...)
}

func (g *RouterGroup) DELETE(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodDelete, path, h, ms...)
}

func (g *RouterGroup) CONNECT(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodConnect, path, h, ms...)
}

func (g *RouterGroup) OPTIONS(path string, h HandlerFunc, ms ...Middleware) {
	g.Handle(http.MethodOptions, path, h, ms...)
}

func (ms Middlewares) Apply(h HandlerFunc) HandlerFunc {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}

	return h
}

func (ms *Middlewares) Append(add ...Middleware) {
	*ms = append(*ms, add...)
}

func (ms Middlewares) Extend(add ...Middleware) Middlewares {
	return append(ms[:len(ms):len(ms)], add...)
}
