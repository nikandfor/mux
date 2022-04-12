package mux

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrSlashRedirect = errors.New("not found, but found ending with slash")
)

func NotFound(c *Context) error {
	c.ResponseWriter.WriteHeader(http.StatusNotFound)
	c.m.dump(c.ResponseWriter, c.m.meth[c.Request.Method], 0, 0)

	return c.Respond(http.StatusNotFound, ErrNotFound)
}

func SlashRedirector(notFound HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		if p := c.URL.Path; p != "" && p[len(p)-1] != '/' {
			h := c.m.Lookup(c.Request.Method, p+"/", nil)
			if h != nil {
				c.ResponseWriter.Header().Set("Location", p+"/")
				c.ResponseWriter.WriteHeader(http.StatusNotFound)

				return ErrSlashRedirect
			}
		}

		return notFound(c)
	}
}
