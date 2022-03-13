package mux

import "net/http"

type (
	Mux struct {
		meth map[string]uint32

		nodes []node
	}

	HandlerFunc func(c *Context) error
)

func New() *Mux {
	return &Mux{}
}

func (m *Mux) FindHandler(meth, path string) HandlerFunc {
	return m.get(meth, path)
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
	err := m.put(meth, path, h)
	if err != nil {
		panic(err)
	}
}

func (m *Mux) GET(path string, h HandlerFunc) {
	m.Handle(http.MethodGet, path, h)
}
