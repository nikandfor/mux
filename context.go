package mux

import (
	"context"
	"net/http"
	"sync"
)

type (
	Context struct {
		context.Context

		http.ResponseWriter
		Request *http.Request
	}
)

var contextPool = sync.Pool{
	New: func() interface{} {
		return new(Context)
	},
}

func getContext() *Context  { return contextPool.Get().(*Context) }
func putContext(c *Context) { contextPool.Put(c) }

func (c *Context) Param(name string) string { return "" }
