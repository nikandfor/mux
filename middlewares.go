package mux

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func Ok(c *Context, w http.ResponseWriter, req *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func Redirect(url string) Handler {
	return RedirectCode(url, http.StatusSeeOther)
}

func RedirectCode(url string, code int) Handler {
	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Location", url)
		w.WriteHeader(code)

		return nil
	}
}

func Alert(msg string) Middleware {
	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		w.Header().Add("X-Alert", msg)

		return nil
	}
}

func Deprecate(t time.Time) Middleware {
	tt := "@" + strconv.FormatInt(t.Unix(), 10)

	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		w.Header().Add("Deprecation", tt)

		return nil
	}
}

func Sunset(t time.Time) Middleware {
	tt := t.UTC().Format(http.TimeFormat)

	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		w.Header().Add("Sunset", tt)

		return nil
	}
}

func Link(url, rel, typ string) Middleware {
	l := fmt.Sprintf(`<%s>;rel="%s";type="%s"`, url, rel, typ)

	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		w.Header().Add("Link", l)

		return nil
	}
}
