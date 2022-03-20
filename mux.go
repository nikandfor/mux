package mux

import (
	"net/http"
	pathpkg "path"
)

type (
	Mux struct {
		RouterGroup

		meth map[string]*node

		NotFound HandlersChain
	}

	RouterGroup struct {
		m *Mux

		basePath string

		hs HandlersChain
	}

	HandlersChain []HandlerFunc

	HandlerFunc func(c *Context) error
)

func New() *Mux {
	m := &Mux{
		meth:     make(map[string]*node),
		NotFound: HandlersChain{NotFound},
	}

	m.RouterGroup = RouterGroup{
		m:        m,
		basePath: "/",
	}

	return m
}

func (g *RouterGroup) Use(hs ...HandlerFunc) {
	if len(hs) == 0 {
		return
	}

	g.hs = append(g.hs, hs...)
}

func (g *RouterGroup) Group(path string, hs ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		m:        g.m,
		basePath: g.basePath + path,
		hs:       append(g.hs[:len(g.hs):len(g.hs)], hs...),
	}
}

func (g *RouterGroup) Handle(meth, path string, hs ...HandlerFunc) {
	hc := make(HandlersChain, len(g.hs)+len(hs))
	copy(hc, g.hs)
	copy(hc[len(g.hs):], hs)

	g.m.handle(meth, g.basePath+path, hc)
}

func (m *Mux) Lookup(meth, path string, c *Context) (h HandlersChain) {
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

func (m *Mux) handle(meth, path string, hs []HandlerFunc) {
	if path == "" || path[0] != '/' {
		panic("bad path")
	}

	cpath := pathpkg.Clean(path)
	if path[len(path)-1] == '/' {
		cpath += "/"
	}

	root := m.meth[meth]
	if root == nil {
		root = &node{}
		m.meth[meth] = root
	}

	node := m.put(root, cpath)

	if node.h != nil {
		panic("path collision: " + cpath)
	}

	node.h = hs
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	defer FreeContext(c)

	c.m = m

	c.hc = m.Lookup(req.Method, req.URL.Path, c)

	if c.hc == nil {
		c.hc = m.RouterGroup.hs
		c.hc2 = m.NotFound
	}

	_ = c.Next()
}

func (g *RouterGroup) GET(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodGet, path, hs...)
}

func (g *RouterGroup) HEAD(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodHead, path, hs...)
}

func (g *RouterGroup) POST(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodPost, path, hs...)
}

func (g *RouterGroup) PUT(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodPut, path, hs...)
}

func (g *RouterGroup) PATCH(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodPatch, path, hs...)
}

func (g *RouterGroup) DELETE(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodDelete, path, hs...)
}

func (g *RouterGroup) CONNECT(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodConnect, path, hs...)
}

func (g *RouterGroup) OPTIONS(path string, hs ...HandlerFunc) {
	g.Handle(http.MethodOptions, path, hs...)
}
