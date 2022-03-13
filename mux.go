package mux

type (
	Mux struct {
		p []page
		f []HandlerFunc
	}

	HandlerFunc func(c *Context) error
)

func (m *Mux) handle(meth, path string, h HandlerFunc) {
	mp := m.put(meth, 0, 0)

	if m.p[mp].val == none {
		ch := m.new()
		m.p[mp].val = ch

		m.p[ch].pref = path
	}

	page := m.put(path, 0, m.p[mp].val)
	if m.p[page].val != -1 {
		panic("handlers collision")
	}

	m.p[page].val = int32(len(m.f))
	m.f = append(m.f, h)
}

func (m *Mux) match(meth, path string, c *Context) HandlerFunc {
	mroot := m.get(meth, 0, 0)
	if mroot == none || m.p[mroot].val == none {
		return nil
	}

	page := m.put(path, 0, m.p[mroot].val)
	if page == none || m.p[page].val == none {
		return nil
	}

	return m.f[m.p[page].val]
}
