package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikandfor/assert"
)

func TestMuxFindHandler(t *testing.T) {
	var m Mux

	ok := false

	m.GET("/progs", func(c *Context) error {
		ok = true
		return nil
	})

	h := m.FindHandler("GET", "/progs")
	if assert.True(t, h != nil, "no handler") {
		h(nil)
		assert.True(t, ok, "called")
	}
}

func TestMuxWrite(t *testing.T) {
	var m Mux

	m.GET("/progs", func(c *Context) error {
		_, err := io.WriteString(c, c.Request.RequestURI)
		return err
	})

	resp := httptest.NewRecorder()

	m.ServeHTTP(resp, &http.Request{
		Method:     http.MethodGet,
		RequestURI: "/progs",
	})

	assert.Equal(t, "/progs", resp.Body.String())
}

func TestMuxStatic(t *testing.T) {
	var m Mux

	h := func(c *Context) error {
		_, _ = io.WriteString(c, c.Request.RequestURI)
		return nil
	}

	for i, tc := range staticRoutes {
		//	fmt.Printf("ADD HANDLER %v %v\n", tc.method, tc.path)

		m.Handle(tc.method, tc.path, h)

		for _, tc := range staticRoutes[:i+1] {
			resp := httptest.NewRecorder()

			m.ServeHTTP(resp, &http.Request{
				Method:     tc.method,
				RequestURI: tc.path,
			})

			assert.Equal(t, tc.path, resp.Body.String())

			if t.Failed() {
				break
			}
		}

		if t.Failed() {
			break
		}

		//	fmt.Printf("routes dump\n%s\n", m.dumpString(m.meth[http.MethodGet]))
	}

	if t.Failed() {
		t.Logf("routes dump\n%s", m.dumpString(m.meth[http.MethodGet]))
	}
}

func BenchmarkMuxStatic(b *testing.B) {
	b.ReportAllocs()

	var m Mux

	h := func(c *Context) error {
		//	_, err := io.WriteString(c, c.Request.RequestURI)
		return nil
	}

	for _, tc := range staticRoutes {
		m.Handle(tc.method, tc.path, h)
	}

	b.ResetTimer()

	resp := httptest.NewRecorder()
	req := http.Request{}

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			//	resp.Body.Reset()
			req.RequestURI = tc.path

			m.ServeHTTP(resp, &req)
		}
	}
}

func BenchmarkMuxStaticWrite(b *testing.B) {
	b.ReportAllocs()

	var m Mux

	h := func(c *Context) error {
		_, _ = io.WriteString(c, c.Request.RequestURI)
		return nil
	}

	for _, tc := range staticRoutes {
		m.Handle(tc.method, tc.path, h)
	}

	b.ResetTimer()

	resp := httptest.NewRecorder()
	req := http.Request{}

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			resp.Body.Reset()

			req.RequestURI = tc.path

			m.ServeHTTP(resp, &req)
		}
	}
}
