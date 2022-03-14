package mux

import (
	"net/http"
	"testing"

	"github.com/nikandfor/assert"
)

func TestMux(t *testing.T) {
	var m Mux

	h := m.match("GET", "/", nil)
	assert.Nil(t, h)

	var res string
	f := func(c *Context) error {
		res = c.Request.RequestURI
		return nil
	}

	want := func(t *testing.T, path string, ps ...string) {
		t.Helper()

		c := &Context{}

		h := m.match("GET", path, c)
		if !assert.True(t, h != nil, path) {
			return
		}
		_ = h(&Context{Request: &http.Request{RequestURI: path}})
		assert.Equal(t, path, res)

		var pp Params
		for i := 0; i < len(ps); i += 2 {
			pp = append(pp, Param{Name: ps[i], Value: ps[i+1]})
		}

		assert.Equal(t, pp, c.Params)
	}

	dontWant := func(t *testing.T, path string) {
		t.Helper()

		h := m.match("GET", path, nil)
		assert.True(t, h == nil)
	}

	defer func() {
		t.Logf("dump GET:\n%s", m.dumpString(m.meth["GET"]))
	}()

	//

	m.handle("GET", "/", f)
	m.handle("GET", "/docs", f)

	dontWant(t, "/qwe")
	want(t, "/")
	want(t, "/docs")

	if t.Failed() {
		return
	}

	m.handle("GET", "/post/{post:\\d+}", f)

	want(t, "/post/1234", "post", "1234")

	if t.Failed() {
		return
	}

	m.handle("GET", "/post/latest", f)

	want(t, "/post/1234", "post", "1234")
	want(t, "/post/latest")

	if t.Failed() {
		return
	}

	m.handle("GET", "/post/laser", f)

	want(t, "/post/1234", "post", "1234")
	want(t, "/post/latest")
	want(t, "/post/laser")
	dontWant(t, "/post/lord")

	if t.Failed() {
		return
	}

	m.handle("GET", "/static/*path", f)
	m.handle("GET", "/users/:user/name", f)
	m.handle("GET", "/users/{user}/profile", f)
	m.handle("GET", "/users/{user}*sub", f)

	want(t, "/static/file", "path", "file")
	want(t, "/static/path/to/file.json", "path", "path/to/file.json")
	want(t, "/users/dan/name", "user", "dan")
	want(t, "/users/dan/profile", "user", "dan")
	want(t, "/users/dan/any/other/path", "user", "dan", "sub", "/any/other/path")

	if t.Failed() {
		return
	}

	m.handle("GET", "/events/:id", f)
	m.handle("GET", "/events/:id/subscribe", f)

	want(t, "/events/aaa", "id", "aaa")
	want(t, "/events/aaa/subscribe", "id", "aaa")

	if t.Failed() {
		return
	}
}
