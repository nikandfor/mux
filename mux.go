package mux

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

type (
	Mux struct {
		Router

		ServeMux http.ServeMux

		// Middlewares that run regardless of whether a handler was found or not.
		Middlewares []Middleware
		NotFound    Handler

		// OnHandle is called each time handler is registered. If set.
		OnHandle func(pattern string, h Handler, ms []Middleware)
	}

	Router struct {
		m *Mux

		Path string

		middlewares []Middleware
	}

	Handler    = func(c *Context, w http.ResponseWriter, req *http.Request) error
	Middleware = Handler

	contextKey struct{}
)

func New() *Mux {
	m := &Mux{}
	m.Router.m = m

	return m
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	c := &Context{
		Context:   ctx,
		Timestamp: time.Now(),

		Mux: m,

		pre:   m.Middlewares,
		route: m.route,
	}

	ctx = context.WithValue(ctx, contextKey{}, c)
	reqctx := req.WithContext(ctx)

	_ = c.Next(w, reqctx) // error should be handled by middlewares
}

func (m *Mux) route(c *Context, w http.ResponseWriter, req *http.Request) error {
	_, pattern := m.ServeMux.Handler(req)
	if pattern == "" {
		c.Break()
		return notFound(c, w, req, m.NotFound)
	}

	m.ServeMux.ServeHTTP(w, req)

	return nil
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

	if f := r.m.OnHandle; f != nil {
		f(p, h, handlers[:len(handlers)-1])
	}

	r.m.ServeMux.HandleFunc(p, func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		c := ContextFrom(ctx)

		c.handlers = handlers
	})
}

func WrapStdHandler(h http.Handler) Handler {
	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		h.ServeHTTP(w, req)
		return nil
	}
}

func WrapStdHandlerFunc(f http.HandlerFunc) Handler {
	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		f(w, req)
		return nil
	}
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

func notFound(c *Context, w http.ResponseWriter, req *http.Request, f Handler) error {
	if f != nil {
		return f(c, w, req)
	}

	http.NotFound(w, req)

	return nil
}
