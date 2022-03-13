package mux

import (
	"testing"

	"github.com/nikandfor/assert"
)

func TestMux(t *testing.T) {
	var m Mux

	h := m.match("GET", "/", nil)
	assert.Nil(t, h)

	var ok1, ok2 bool
	f1 := func(c *Context) error {
		ok1 = true
		return nil
	}
	f2 := func(c *Context) error {
		ok2 = true
		return nil
	}

	m.handle("GET", "/v0", f1)
	m.handle("GET", "/v0/docs", f2)

	h = m.match("GET", "/", nil)
	assert.Nil(t, h)

	h = m.match("GET", "/v0", nil)
	if assert.True(t, h != nil) {
		assert.False(t, ok1)
		h(nil)
		assert.True(t, ok1)
	}

	h = m.match("GET", "/v0/docs", nil)
	if assert.True(t, h != nil) {
		assert.False(t, ok2)
		h(nil)
		assert.True(t, ok2)
	}

	//	t.Logf("dump:\n%s", m.dumpString(m.root))
	t.Logf("dump GET:\n%s", m.dumpString(m.meth["GET"]))
}
