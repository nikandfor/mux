package mux

import (
	"testing"
	"unsafe"

	"github.com/nikandfor/assert"
)

func TestTree(t *testing.T) {
	var m Mux

	f1 := func(*Context) error { return nil }
	f2 := func(*Context) error { return nil }

	err := m.put("GET", "/dog", f1)
	assert.NoError(t, err)

	t.Logf("dump %v  (%d nodes)\n%s", m.meth, len(m.nodes), m.dumpString(m.meth["GET"]))

	assert.Equal(t, fn(f1), fn(m.get("GET", "/dog")), "/dog")

	err = m.put("GET", "/dolly", f2)
	assert.NoError(t, err)

	t.Logf("dump %v  (%d nodes)\n%s", m.meth, len(m.nodes), m.dumpString(m.meth["GET"]))

	assert.Equal(t, fn(f1), fn(m.get("GET", "/dog")), "/dog")
	assert.Equal(t, fn(f2), fn(m.get("GET", "/dolly")), "/dolly")
	assert.Equal(t, fn(nil), fn(m.get("GET", "/dolly1")), "/dolly1")
}

func TestSearch(t *testing.T) {
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{}, len: 0}, '1'))
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{'5'}, len: 1}, '5'))
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{'6'}, len: 1}, '5'))
	assert.Equal(t, 1, search(&node{x: [pagesize]uint32{'4'}, len: 1}, '5'))
	assert.Equal(t, 1, search(&node{x: [pagesize]uint32{'4', '5'}, len: 2}, '5'))
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{'5', '6'}, len: 2}, '5'))
	assert.Equal(t, 1, search(&node{x: [pagesize]uint32{'4', '5', '6'}, len: 3}, '5'))
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '2'))
	assert.Equal(t, 0, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '3'))
	assert.Equal(t, 1, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '4'))
	assert.Equal(t, 1, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '5'))
	assert.Equal(t, 2, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '6'))
	assert.Equal(t, 2, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '7'))
	assert.Equal(t, 3, search(&node{x: [pagesize]uint32{'3', '5', '7'}, len: 3}, '8'))
}

func BenchmarkTreeStatic(b *testing.B) {
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

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	root := m.meth["GET"]

	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			m.nodeGet(tc.path, 0, root)
		}
	}
}

func fn(f func(*Context) error) uintptr {
	if f == nil {
		return 0
	}

	return **(**uintptr)(unsafe.Pointer(&f))
}
