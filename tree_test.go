package mux

import (
	"fmt"
	"testing"

	"github.com/nikandfor/assert"
)

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

	fmt.Printf("dump\n%s\n", m.dumpString(m.meth["GET"]))

	//	assert.Equal(t, m.root, m.get("/dog", 0, m.root), "/dog")
	hh := m.get("", "/dog", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dog", res)
	}

	fmt.Printf("===\n")

	m.handle("", "/dolly", h("/dolly"))

	fmt.Printf("dump\n%s\n", m.dumpString(m.meth["GET"]))

	hh = m.get("", "/dog", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dog", res)
	}

	hh = m.get("", "/dolly", nil)
	if assert.True(t, hh != nil && hh(nil) == nil) {
		assert.Equal(t, "/dolly", res)
	}

	hh = m.get("", "/dolly1", nil)
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
		fmt.Printf("ADD HANDLER %v %v\n", tc.method, tc.path)

		m.handle("", tc.path, h(tc.path))

		for _, tc := range staticRoutes[:i+1] {
			h := m.get("", tc.path, nil)
			if assert.True(t, h != nil && h(nil) == nil, tc.path) {
				assert.Equal(t, tc.path, res)
			}

			if t.Failed() {
				break
			}
		}

		fmt.Printf("routes dump\n%s\n", m.dumpString(m.meth[""]))

		if t.Failed() {
			break
		}
	}

	//if t.Failed() {
	t.Logf("routes dump\n%s", m.dumpString(m.meth[""]))
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
	hh := m.get("", staticRoutes[l].path, nil)
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
			m.get("", tc.path, nil)
		}
	}
}
