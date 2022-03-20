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
		basePath: pathpkg.Join(g.basePath, path),
		hs:       append(g.hs[:len(g.hs):len(g.hs)], hs...),
	}
}

func (g *RouterGroup) Handle(meth, path string, hs ...HandlerFunc) {
	hc := make(HandlersChain, len(g.hs)+len(hs))
	copy(hc, g.hs)
	copy(hc[len(g.hs):], hs)

	g.m.handle(meth, pathpkg.Join(g.basePath, path), hc)
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
	path = pathpkg.Clean(path)

	root := m.meth[meth]
	if root == nil {
		root = &node{}
		m.meth[meth] = root
	}

	node := m.put(root, path)

	if node.h != nil {
		panic("path collision: " + path)
	}

	node.h = hs
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	defer FreeContext(c)

	c.hc = m.Lookup(req.Method, req.URL.Path, c)

	if c.hc == nil {
		c.hc = m.RouterGroup.hs
		c.hc2 = m.NotFound
	}

	_ = c.Next()
}
