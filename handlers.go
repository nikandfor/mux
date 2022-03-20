package mux

import "net/http"

func NotFound(c *Context) error {
	http.NotFound(c.ResponseWriter, c.Request)

	c.m.dump(c.ResponseWriter, c.m.meth[c.Request.Method], 0, 0)

	return nil
}
