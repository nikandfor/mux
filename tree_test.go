package mux

import (
	"fmt"
	"testing"

	"github.com/nikandfor/assert"
)

func TestSearch(t *testing.T) {
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{}, len: 0}, '1'))
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{'5'}, len: 1}, '5'))
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{'6'}, len: 1}, '5'))
	assert.Equal(t, int8(1), search(&page{x: [pagesize]int32{'4'}, len: 1}, '5'))
	assert.Equal(t, int8(1), search(&page{x: [pagesize]int32{'4', '5'}, len: 2}, '5'))
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{'5', '6'}, len: 2}, '5'))
	assert.Equal(t, int8(1), search(&page{x: [pagesize]int32{'4', '5', '6'}, len: 3}, '5'))
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '2'))
	assert.Equal(t, int8(0), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '3'))
	assert.Equal(t, int8(1), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '4'))
	assert.Equal(t, int8(1), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '5'))
	assert.Equal(t, int8(2), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '6'))
	assert.Equal(t, int8(2), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '7'))
	assert.Equal(t, int8(3), search(&page{x: [pagesize]int32{'3', '5', '7'}, len: 3}, '8'))
}

func TestTree(t *testing.T) {
	var m Mux

	_ = m.putVal("/dog", 0, 1)

	fmt.Printf("dump (%d nodes)\n%s\n", len(m.p), m.dumpString(0))

	assert.Equal(t, int32(1), m.getVal("/dog", 0), "/dog")

	fmt.Printf("===\n")

	_ = m.putVal("/dolly", 0, 2)

	fmt.Printf("dump (%d nodes)\n%s\n", len(m.p), m.dumpString(0))

	assert.Equal(t, int32(1), m.getVal("/dog", 0), "/dog")
	assert.Equal(t, int32(2), m.getVal("/dolly", 0), "/dolly")
	assert.Equal(t, none, m.getVal("/dolly1", 0), "/dolly1")
}

func TestTreeStatic(t *testing.T) {
	var m Mux

	for i, tc := range staticRoutes {
		fmt.Printf("ADD HANDLER %v %v\n", tc.method, tc.path)

		m.putVal(tc.path, 0, int32(i))

		for j, tc := range staticRoutes[:i+1] {
			ii := m.getVal(tc.path, 0)

			assert.Equal(t, int32(j), ii, tc.path)

			if t.Failed() {
				break
			}
		}

		fmt.Printf("routes dump\n%s\n", m.dumpString(0))

		if t.Failed() {
			break
		}
	}

	//	if t.Failed() {
	//		t.Logf("routes dump\n%s", m.dumpString(0))
	//	}
}

func BenchmarkTreeStatic(b *testing.B) {
	b.ReportAllocs()

	var m Mux

	for i, tc := range staticRoutes {
		m.putVal(tc.path, 0, int32(i))
	}

	b.ResetTimer()

	n := b.N / len(staticRoutes)
	if n == 0 {
		n = 1
	}

	var val int32
	for i := 0; i < n; i++ {
		for _, tc := range staticRoutes {
			val = m.getVal(tc.path, 0)
		}
	}

	assert.Equal(b, int32(len(staticRoutes)-1), val)
}
