package mux

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type (
	Context struct {
		context.Context

		ResponseWriter http.ResponseWriter
		*http.Request

		Params

		QueryValues url.Values

		Responder
		Encoder

		m *Mux

		i   int
		hc  HandlersChain
		hc2 HandlersChain
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
	c.Context = nil
	c.ResponseWriter = nil
	c.Request = nil

	for i := range c.Params {
		c.Params[i] = Param{}
	}

	c.Params = c.Params[:0]
	c.QueryValues = nil

	c.m = nil

	c.i = 0
	c.hc = nil
	c.hc2 = nil

	contextPool.Put(c)
}

func (c *Context) Next() (err error) {
	for c.i < len(c.hc)+len(c.hc2) {
		c.i++
		if c.i <= len(c.hc) {
			err = c.hc[c.i-1](c)
		} else {
			err = c.hc2[c.i-1-len(c.hc)](c)
		}
		if err != nil {
			return err
		}
	}

	return nil
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

func (c *Context) ClientIP() (ip string) {
	ip, _, _ = net.SplitHostPort(c.Request.RemoteAddr)
	return
}
