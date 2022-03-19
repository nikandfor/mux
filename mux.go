package mux

import (
	pathpkg "path"
)

type (
	Mux struct {
		RouterGroup

		meth map[string]*node
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
	m := &Mux{}

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
	g.m.handle(meth, pathpkg.Join(g.basePath, path), append(g.hs[:len(g.hs):len(g.hs)], hs...))
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

	if m.meth == nil {
		m.meth = make(map[string]*node)
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

	node.h = hs
}
