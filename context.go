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
		*http.Request

		Params
	}

	Params []Param

	Param struct {
		Name  string
		Value string
	}
)

var contextPool = sync.Pool{
	New: func() interface{} {
		return new(Context)
	},
}

func NewContext(w http.ResponseWriter, req *http.Request) (c *Context) {
	c = contextPool.Get().(*Context)

	c.Context = req.Context()
	c.ResponseWriter = w
	c.Request = req

	return c
}

func FreeContext(c *Context) {
	c.Params = c.Params[:0]

	contextPool.Put(c)
}

func (ps Params) LookupParam(name string) (string, bool) {
	for _, p := range ps {
		if p.Name == name {
			return p.Value, true
		}
	}

	return "", false
}

func (ps Params) Param(name string) (v string) {
	v, _ = ps.LookupParam(name)
	return
}
