package mux

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func OK(c *Context, w http.ResponseWriter, req *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func Redirect(url string) Handler {
	return RedirectCode(url, http.StatusSeeOther)
}

func RedirectCode(url string, code int) Handler {
	absolute := strings.HasPrefix(url, "/")

	return func(c *Context, w http.ResponseWriter, req *http.Request) error {
		loc := url

		if absolute {
			prefix := req.Header.Get("X-Forwarded-Prefix")
			loc = Join(prefix, loc)
		}

		w.Header().Set("Location", loc)
		w.WriteHeader(code)

		return nil
	}
}

func Alert(format string, args ...any) Middleware {
	msg := fmt.Sprintf(format, args...)

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
