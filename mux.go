package mux

import "net/http"

type (
	Mux struct {
		meth map[string]uint32

		nodes []node
		funcs []HandlerFunc
	}

	HandlerFunc func(c *Context) error
)

func New() *Mux {
	return &Mux{}
}

func (m *Mux) FindHandler(meth, path string) HandlerFunc {
	h := m.get(meth, path)
	if h < 0 {
		return nil
	}

	return m.funcs[h]
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h := m.FindHandler(req.Method, req.RequestURI)
	if h == nil {
		return
	}

	c := getContext()
	defer putContext(c)

	c.Context = req.Context()
	c.ResponseWriter = w
	c.Request = req

	err := h(c)
	_ = err
}

func (m *Mux) Handle(meth, path string, h HandlerFunc) {
	hid := int32(len(m.funcs))
	m.funcs = append(m.funcs, h)

	err := m.put(meth, path, hid)
	if err != nil {
		panic(err)
	}
}

func (m *Mux) GET(path string, h HandlerFunc) {
	m.Handle(http.MethodGet, path, h)
}
