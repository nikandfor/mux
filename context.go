package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type (
	Context struct {
		Context context.Context

		KV map[any]any

		Timestamp time.Time

		Mux *Mux

		pre      []Handler
		route    Handler
		handlers []Handler

		index int
	}
)

func ContextFrom(ctx context.Context) *Context {
	v := ctx.Value(contextKey{})
	c, _ := v.(*Context)

	return c
}

func (c *Context) Next(w http.ResponseWriter, req *http.Request) (err error) {
	for c.index < len(c.pre) {
		index := c.index
		c.index++

		err := c.pre[index](c, w, req)
		if err != nil {
			return err
		}
	}

	if c.index == len(c.pre) {
		err = c.route(c, w, req)
		if err != nil {
			return err
		}
	}

	for c.index < len(c.pre)+len(c.handlers) {
		index := c.index
		c.index++

		err := c.handlers[index-len(c.pre)](c, w, req)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) Break() { c.index = 1e10 }

func (c *Context) SetKV(k, v any) {
	if c.KV == nil {
		c.KV = make(map[any]any)
	}

	c.KV[k] = v
}

func (c *Context) RespondJSON(w http.ResponseWriter, v any) error {
	return c.RespondJSONCode(w, v, http.StatusOK)
}

func (c *Context) RespondJSONCode(w http.ResponseWriter, v any, code int) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	data = append(data, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (c *Context) UnmarshalRequestJSON(req *http.Request, v any) error {
	d := json.NewDecoder(req.Body)

	err := d.Decode(v)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}
