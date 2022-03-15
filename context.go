package mux

import (
	"context"
	"net/http"
	"net/url"
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

		QueryValues url.Values

		Binder

		handlers []HandlerFunc
		index    int
	}

	Params []Param

	Param struct {
		Name  string
		Value string
	}

	Responder interface {
		Respond(code int, msg interface{}) error
	}

	Binder interface {
		Bind(msg interface{}) error
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

	return c
}

func PutContext(c *Context) {
	c.Params = c.Params[:0]

	contextPool.Put(c)
}

func (c *Context) Next() (err error) {
	for c.index < len(c.handlers) {
		c.index++
		err = c.handlers[c.index-1](c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) Query(k string) (v string) {
	if c.QueryValues == nil {
		c.QueryValues = c.Request.URL.Query()
	}

	return c.QueryValues.Get(k)
}

func (c *Context) LookupQuery(k string) (v string, ok bool) {
	v = c.Query(k)
	_, ok = c.QueryValues[k]
	return
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
