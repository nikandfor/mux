package mux

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type (
	Mux struct {
		Router

		ServeMux http.ServeMux

		// Middlewares that are run independent of if handler was found or not.
		Middlewares []Middleware
		NotFound    Handler

		HandlerHook func(pattern string, h Handler, ms []Middleware)
	}

	Router struct {
		m *Mux

		Path string

		middlewares []Middleware
	}

	Handler    = func(c *Context, w http.ResponseWriter, req *http.Request) error
	Middleware = Handler
)

func New() *Mux {
	m := &Mux{
		NotFound: notFound,
	}
	m.Router.m = m

	return m
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, pattern := m.ServeMux.Handler(req)
	if pattern == "" {
		return
	}

	m.ServeMux.ServeHTTP(w, req)
}

func (m *Mux) handle(w http.ResponseWriter, req *http.Request, hs []Handler) {
	ctx := req.Context()

	c := &Context{
		Context: ctx,

		handlers: m.Middlewares,
	}

	err := c.Next(w, req)
	if err != nil {
		return
	}

	c.handlers = hs

	err = c.Next(w, req)
	if err != nil {
		return
	}
}

func (r *Router) Group(path string, ms ...Middleware) *Router {
	return &Router{
		m: r.m,

		Path: Join(r.Path, path),

		middlewares: slices.Concat(r.middlewares, ms),
	}
}

func (r *Router) Use(ms ...Middleware) {
	r.middlewares = append(r.middlewares, ms...)
}

func (r *Router) Handle(pattern string, h Handler, ms ...Middleware) {
	meth, path := splitPattern(pattern)
	path = Join(r.Path, path)
	p := joinPattern(meth, path)

	handlers := slices.Concat(r.middlewares, ms, []Handler{h})

	if f := r.m.HandlerHook; f != nil {
		f(pattern, h, handlers[:len(handlers)-1])
	}

	r.m.ServeMux.HandleFunc(p, func(w http.ResponseWriter, req *http.Request) {
		r.m.handle(w, req, handlers)
	})
}

func Join(elems ...string) string {
	if len(elems) == 0 {
		return ""
	}

	var b strings.Builder

	for _, elem := range elems {
		elem = strings.Trim(elem, "/")
		if elem == "" || elem == "." {
			continue
		}

		_ = b.WriteByte('/')
		_, _ = b.WriteString(elem)
	}

	if strings.HasSuffix(elems[len(elems)-1], "/") {
		_ = b.WriteByte('/')
	}

	return b.String()
}

func joinPattern(meth, path string) string {
	if meth == "" {
		return path
	}

	return fmt.Sprintf("%-4s %s", meth, path)
}

func splitPattern(p string) (meth, path string) {
	i := strings.IndexByte(p, ' ')
	if i == -1 {
		return "", p
	}

	meth = p[:i]

	for i < len(p) && p[i] == ' ' {
		i++
	}

	if i == len(p) {
		path = "."
	} else {
		path = p[i:]
	}

	return meth, path
}

func notFound(c *Context, w http.ResponseWriter, req *http.Request) error {
	http.NotFound(w, req)

	return nil
}
