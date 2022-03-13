package mux

import (
	"testing"

	"github.com/nikandfor/assert"
)

func TestTree(t *testing.T) {
	var m Mux

	err := m.put("GET", "/dog", 1)
	assert.NoError(t, err)

	t.Logf("dump %v  (%d nodes)\n%s", m.meth, len(m.nodes), m.dumpString(m.meth["GET"]))

	assert.Equal(t, int32(1), m.get("GET", "/dog"), "/dog")

	err = m.put("GET", "/dolly", 2)
	assert.NoError(t, err)

	t.Logf("dump %v  (%d nodes)\n%s", m.meth, len(m.nodes), m.dumpString(m.meth["GET"]))

	assert.Equal(t, int32(1), m.get("GET", "/dog"), "/dog")
	assert.Equal(t, int32(2), m.get("GET", "/dolly"), "/dolly")
	assert.Equal(t, int32(-1), m.get("GET", "/dolly1"), "/dolly1")
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
