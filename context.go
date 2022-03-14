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

		Params

		Responder
		Encoder
	}

	Params []Param

	Param struct {
		Name  string
		Value string
	}

	Responder interface {
		Respond(code int, msg interface{}) error
	}

	Encoder interface {
		Encode(msg interface{}) error
	}
)

var contextPool = sync.Pool{
	New: func() interface{} {
		return new(Context)
	},
}

func GetContext(w http.ResponseWriter, req *http.Request) (c *Context) {
	c = contextPool.Get().(*Context)

	c.Context = req.Context()
	c.ResponseWriter = w
	c.Request = req
	c.Params = c.Params[:0]

	return c
}

func PutContext(c *Context) { contextPool.Put(c) }

func (ps Params) Param(name string) string {
	for _, p := range ps {
		if p.Name == name {
			return p.Value
		}
	}

	return ""
}
