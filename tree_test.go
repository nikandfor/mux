package mux

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/nikandfor/assert"
)

var dumpAll = flag.Bool("dump-all", false, "dump tree while tested")

func TestTree(t *testing.T) {
	var res string
	h := func(p string) HandlerFunc {
		return func(c *Context) error {
			res = p
			return nil
		}
	}

	var m Mux

	m.handle("", "/dog", h("/dog"))

	fmt.Printf("dump\n%s\n", m.dumpString(m.meth[""]))

	//	assert.Equal(t, m.root, m.get("/dog", 0, m.root), "/dog")
	hh := m.match("", "/dog", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dog", res)
	}

	fmt.Printf("===\n")

	m.handle("", "/dolly", h("/dolly"))

	fmt.Printf("dump\n%s\n", m.dumpString(m.meth[""]))

	hh = m.match("", "/dog", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dog", res)
	}

	hh = m.match("", "/dolly", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dolly", res)
	}

	hh = m.match("", "/dolly1", nil)
	assert.True(t, hh == nil)
}

func TestTreeStatic(t *testing.T) {
	var m Mux

	var res string
	h := func(p string) HandlerFunc {
		return func(c *Context) error {
			res = p
			return nil
		}
	}

	for i, tc := range staticRoutes {
		if *dumpAll {
			fmt.Printf("ADD HANDLER %v %v\n", tc.method, tc.path)
		}

		m.handle("", tc.path, h(tc.path))

		for _, tc := range staticRoutes[:i+1] {
			h := m.match("", tc.path, nil)
			if assert.True(t, h != nil && h(nil) == nil, tc.path) {
				assert.Equal(t, tc.path, res)
			}

			if t.Failed() {
				break
			}
		}

		if *dumpAll {
			fmt.Printf("routes dump\n%s\n", m.dumpString(m.meth[""]))
		}

		if t.Failed() {
			break
		}
	}

	//if t.Failed() {
	t.Logf("routes dump\n%s", m.dumpString(m.meth[""]))
	//}
}

func TestTreeGithub(t *testing.T) {
	var m Mux

	var hmeth, hpath, body string
	f := func(meth, path string) func(c *Context) error {
		return func(c *Context) error {
			hmeth = meth
			hpath = path
			body = c.Request.RequestURI
			return nil
		}
	}

	want := func(t *testing.T, meth, path string) {
		c := &Context{Request: &http.Request{
			RequestURI: path,
		}}

		h := m.match(meth, strings.ReplaceAll(path, ":", ""), c)
		if !assert.True(t, h != nil, "%v %v", meth, path) {
			return
		}

		_ = h(c)
		assert.Equal(t, meth, hmeth, "tree path mismatch")
		assert.Equal(t, path, hpath, "tree path mismatch")
		assert.Equal(t, path, body, "request processing mismatch")

		assert.Equal(t, strings.Count(path, ":"), len(c.Params))

		for _, p := range c.Params {
			assert.Equal(t, p.Name, p.Value)
		}
	}

	for i, tc := range githubAPI {
		if *dumpAll {
			fmt.Printf("ADD HANDLER %v %v\n", tc.method, tc.path)
		}

		m.handle(tc.method, tc.path, f(tc.method, tc.path))

		for _, tc := range githubAPI[:i+1] {
			want(t, tc.method, tc.path)

			if t.Failed() {
				break
			}
		}

		if *dumpAll {
			fmt.Printf("routes dump\n%s\n", m.dumpString(m.meth[""]))
		}

		if t.Failed() {
			break
		}
	}

	//if t.Failed() {
	for meth, p := range m.meth {
		t.Logf("routes dump  %v\n%s", meth, m.dumpString(p))
	}
	//}
}

func BenchmarkTreeStatic(b *testing.B) {
	b.ReportAllocs()

	var res string
	h := func(p string) HandlerFunc {
		return func(c *Context) error {
			res = p
			return nil
		}
	}

	var m Mux

	for _, tc := range staticRoutes {
		m.handle("", tc.path, h(tc.path))
	}

	l := len(staticRoutes) - 1
	hh := m.match("", staticRoutes[l].path, nil)
	if assert.True(b, hh != nil && hh(nil) == nil) {
		assert.Equal(b, staticRoutes[l].path, res)
	}

	b.ResetTimer()

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			m.match("", tc.path, nil)
		}
	}
}

func BenchmarkHttpRouterStatic(b *testing.B) {
	b.ReportAllocs()

	var res string
	h := func(p string) httprouter.Handle {
		return func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
			res = p
		}
	}

	var m httprouter.Router

	for _, tc := range staticRoutes {
		m.Handle("", tc.path, h(tc.path))
	}

	l := len(staticRoutes) - 1
	hh, _, _ := m.Lookup("", staticRoutes[l].path)
	if assert.True(b, hh != nil) {
		hh(nil, nil, nil)
		assert.Equal(b, staticRoutes[l].path, res)
	}

	b.ResetTimer()

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			m.Lookup("", tc.path)
		}
	}
}

func BenchmarkBothStatic(b *testing.B) {
	b.ReportAllocs()

	var m Mux
	var r httprouter.Router

	for _, tc := range staticRoutes {
		m.handle("", tc.path, nil)
		r.Handle("", tc.path, nil)
	}

	b.ResetTimer()

	//	Comps = 0
	//	httprouter.Comps = 0

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			m.match("", tc.path, nil)
			r.Lookup("", tc.path)
		}
	}

	//	b.Logf("comparisons %d %d", Comps, httprouter.Comps)
}

func TestBothStatic(t *testing.T) {
	var res string
	h1 := func(p string) HandlerFunc {
		return func(c *Context) error {
			res = p
			return nil
		}
	}
	h2 := func(p string) httprouter.Handle {
		return func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
			res = p
		}
	}

	var m Mux
	var r httprouter.Router

	for _, tc := range staticRoutes {
		m.handle("", tc.path, h1(tc.path))
		r.Handle("", tc.path, h2(tc.path))
	}

	path := "/articles/wiki/final-noerror.go"

	//	Comps = 0
	//	httprouter.Comps = 0

	hh1 := m.match("", path, nil)
	hh2, _, _ := r.Lookup("", path)

	res = ""
	hh1(nil)
	assert.Equal(t, path, res)

	res = ""
	hh2(nil, nil, nil)
	assert.Equal(t, path, res)

	//	t.Logf("comparisons %d %d", Comps, httprouter.Comps)
}

func BenchmarkTreeGithub(b *testing.B) {
	b.ReportAllocs()

	var m Mux

	m.NotFoundHandler = func(c *Context) error {
		assert.Fail(b, "NOT FOUND: %v %v", c.Request.Method, c.Request.RequestURI)
		return nil
	}

	f := func(meth, path string) func(c *Context) error {
		return func(c *Context) error {
			io.WriteString(c.ResponseWriter, c.Request.RequestURI)
			return nil
		}
	}

	replace := make([]string, len(githubAPI))

	for i, tc := range githubAPI {
		m.handle(tc.method, tc.path, f(tc.method, tc.path))
		replace[i] = strings.ReplaceAll(tc.path, ":", "")
	}

	b.ResetTimer()

	n := b.N / len(githubAPI)
	if n == 0 {
		n = 1
	}

	w := httptest.NewRecorder()
	req := &http.Request{}

	for i := 0; i < n; i++ {
		for j, tc := range githubAPI {
			w.Body.Reset()

			req.Method = tc.method
			req.RequestURI = tc.path

			func() {
				c := GetContext(w, req)
				defer PutContext(c)

				h := m.match(req.Method, replace[j], c)
				if h == nil {
					h = m.NotFoundHandler
				}

				_ = h(c)
			}()
		}
	}
}
