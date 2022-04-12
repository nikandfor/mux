package mux

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
)

type (
	ResponderFunc func(code int, msg interface{}) error
	EncoderFunc   func(msg interface{}) error

	StatusError struct {
		Status int
		Header http.Header
		Err    error
	}
)

//func (f ResponderFunc) Respond(code int, msg interface{}) error { return f(code, msg) }
//func (f EncoderFunc) Encode(msg interface{}) error              { return f(msg) }

func BasicResponder(next HandlerFunc) HandlerFunc {
	return func(c *Context) (err error) {
		c.Respond = func(code int, msg interface{}) error {
			//	tlog.Printw("respond", "code", code, "msg", msg, "msg_type", tlog.FormatNext("%T"), msg, "from", loc.Callers(1, 5))
			c.ResponseWriter.WriteHeader(code)
			return c.Encode(msg)
		}

		return next(c)
	}
}

func OnceResponder(next HandlerFunc) HandlerFunc {
	return func(c *Context) (err error) {
		var once int32

		prev := c.Respond

		c.Respond = func(code int, msg interface{}) error {
			if !atomic.CompareAndSwapInt32(&once, 0, 1) {
				return errors.New("already answered")
			}

			return prev(code, msg)
		}

		return next(c)
	}
}

func ErrorResponder(next HandlerFunc) HandlerFunc {
	return func(c *Context) (err error) {
		prev := c.Respond

		cur := func(code int, msg interface{}) error {
			err, ok := msg.(error)
			if !ok {
				return prev(code, msg)
			}

			var se StatusError
			if errors.As(err, &se) {
				code = se.Status

				for k, v := range se.Header {
					for _, v := range v {
						c.ResponseWriter.Header().Add(k, v)
					}
				}
			}

			return prev(code, map[string]interface{}{"error": err.Error()})
		}

		c.Respond = cur

		err = next(c)
		if err != nil {
			_ = cur(http.StatusInternalServerError, err)
		}

		return err
	}
}

func MultiEncoder(e map[string]EncoderFunc, def EncoderFunc) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) (err error) {
			for _, l := range c.Request.Header["Accept"] {
				for _, ct := range strings.Split(l, ",") {
					ct = strings.TrimSpace(ct)
					if p := strings.IndexByte(ct, ';'); p != -1 {
						ct = ct[:p]
					}

					ef, ok := e[ct]
					if ok {
						c.ResponseWriter.Header().Set("Content-Type", ct)
						c.Encode = ef

						return next(c)
					}
				}
			}

			c.Encode = def

			return next(c)
		}
	}
}

func JSONEncoder(next HandlerFunc) HandlerFunc {
	return func(c *Context) (err error) {
		c.ResponseWriter.Header().Set("Content-Type", "application/json")
		c.Encode = json.NewEncoder(c.ResponseWriter).Encode

		return next(c)
	}
}

func (e StatusError) Error() string {
	return e.Err.Error()
}

func (e StatusError) Unwrap() error {
	return e.Err
}
