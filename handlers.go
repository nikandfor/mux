package mux

import "net/http"

func NotFound(c *Context) error {
	http.NotFound(c.ResponseWriter, c.Request)

	return nil
}
