package mux

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type (
	node struct {
		pref string

		h HandlersChain

		s map[byte]*node

		wild *node

		// param

		name string
		eat  eater
	}

	eater func(path string) (string, bool)

	wild struct {
		n      *node
		path   string
		params int
	}
)

var spaces = "                                                                                                                     "

func (m *Mux) get(root *node, path string, c *Context) (n *node) {
	n = root

	var wwbuf [20]wild
	ww := wwbuf[:0]

	params := 0

	for {
		//	fmt.Printf("get  %v from %p %+v\n", path, n, n)

		if n != nil && len(n.pref) <= len(path) && n.pref == path[:len(n.pref)] {
			if len(n.pref) == len(path) {
				return n
			}

			path = path[len(n.pref):]

			sub := n.s[path[0]]

			if n.wild != nil {
				//	fmt.Printf("save wild %p %+v\n", n.wild, n.wild)
				if n.wild.name == "" {
					panic("bad wild")
				}

				ww = append(ww, wild{
					n:      n.wild,
					path:   path,
					params: params,
				})
			}

			n = sub
		} else {
			if len(ww) == 0 {
				return nil
			}

			l := len(ww) - 1
			n, path, params = ww[l].n, ww[l].path, ww[l].params
			ww = ww[:l]

			//	fmt.Printf("param %+v  %v\n", n, path)

			val, ok := n.eat(path)
			if !ok {
				n = nil
				continue
			}

			if c != nil {
				c.Params = append(c.Params[:params], Param{
					Name:  n.name,
					Value: val,
				})

				params++
			}

			path = path[len(val):]
		}
	}
}

func (m *Mux) put(root *node, path string) (n *node) {
	n = root

	for i := 0; i < len(path); {
		st := i
		i = indexAny(path, i+1, "{:*")

		n = m.putStatic(n, path[st:i])

		if i == len(path) {
			break
		}

		st = i + 1

		var name string
		var eat eater

		switch path[i] {
		case '*':
			i = len(path)
			name = path[st:]
			eat = eatFull
		case ':':
			i = indexAny(path, st, ":*{/")
			if i != len(path) && path[i] != '/' {
				panic("bad pattern: colon eats all the segment")
			}

			name = path[st:i]
			eat = eatSeg
		default:
			i = indexAny(path, st, ":}")

			if i == len(path) {
				panic("bad pattern: no }")
			}

			name = path[st:i]

			if path[i] == ':' {
				st = i + 1
				i = index(path, st, '}')
				eat = eatRE(path[st:i])
			}

			if i == len(path) || path[i] != '}' {
				panic("bad pattern: no }")
			}

			if eat == nil {
				st = i + 1
				end := index(path, st, '/')
				eat = eatUntil(path[st:end])
			}

			i++
		}

		if n.wild != nil && name != n.wild.name {
			panic("wildcard collision: " + path)
		} else if n.wild == nil {
			n.wild = &node{
				name: name,
				eat:  eat,
			}
		}

		n = n.wild
	}

	return n
}

func (m *Mux) putStatic(root *node, path string) (n *node) {
	if root.pref == "" && root.name == "" {
		root.pref = path

		return root
	}

	n = root

	for {
		c := common(n.pref, path)

		if c != len(n.pref) {
			sub := &node{
				pref: n.pref[c:],
				h:    n.h,
				s:    n.s,
				wild: n.wild,
			}

			n.pref = n.pref[:c]
			n.h = nil
			n.wild = nil

			n.s = map[byte]*node{
				sub.pref[0]: sub,
			}
		}

		if c == len(path) {
			return n
		}

		path = path[c:]

		sub := n.s[path[0]]
		if sub != nil {
			n = sub
			continue
		}

		sub = &node{
			pref: path,
		}

		if n.s == nil {
			n.s = make(map[byte]*node)
		}

		n.s[path[0]] = sub

		return sub
	}
}

func (m *Mux) dumpString(n *node) string {
	if n == nil {
		return "<nil>"
	}

	var b bytes.Buffer

	m.dump(&b, n, 0, 0)

	return b.String()
}

func (m *Mux) dump(w io.Writer, n *node, d, st int) {
	const maxdepth = 10
	if d > maxdepth {
		fmt.Fprintf(w, "%v...\n", spaces[:2*d])
		return
	}

	pref := n.pref
	if n.name != "" {
		pref = "{" + n.name + "}" + n.pref
	}

	valPad := 36 - st - len(pref)
	if valPad < 0 {
		valPad = 0
	}

	fmt.Fprintf(w, "   dump %v%d%v  pref %v%q%v  h %p  %v\n", spaces[:2*d], d, spaces[:2*(maxdepth-d)], spaces[:st], pref, spaces[:valPad], n.h, n.ll())
	for _, sub := range n.s {
		m.dump(w, sub, d+1, st+len(pref))
	}

	if n.wild != nil {
		m.dump(w, n.wild, d+1, st+len(pref))
	}
}

func (n *node) ll() string {
	var b bytes.Buffer

	b.WriteByte('[')

	for k, s := range n.s {
		fmt.Fprintf(&b, " %q:%d", k, countHandlers(s))
	}

	b.WriteByte(']')

	return b.String()
}

func index(s string, st int, c byte) (i int) {
	for i = st; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}

	return i
}

func indexAny(s string, st int, c string) (i int) {
	for i = st; i < len(s); i++ {
		for _, c := range []byte(c) {
			if s[i] == c {
				return i
			}
		}
	}

	return i
}

func common(a, b string) (c int) {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}

	_, _ = a[:m], b[:m]

	for c < m && a[c] == b[c] {
		c++
	}

	return c
}

func eatFull(path string) (string, bool) {
	return path, true
}

func eatSeg(path string) (string, bool) {
	return path[:index(path, 0, '/')], true
}

func eatUntil(until string) eater {
	return eater(func(path string) (val string, ok bool) {
		i := index(path, 0, until[0])
		ok = strings.HasPrefix(path[i:], until)
		val = path[:i]
		return
	})
}

func eatRE(pattern string) eater {
	re := regexp.MustCompile(`^` + pattern)

	return eater(func(path string) (val string, ok bool) {
		if !re.MatchString(path) {
			return "", false
		}

		return re.FindString(path), true
	})
}

func countHandlers(n *node) (sum int) {
	if n.h != nil {
		sum++
	}

	for _, s := range n.s {
		sum += countHandlers(s)
	}

	if s := n.wild; s != nil {
		sum += countHandlers(s)
	}

	return
}
