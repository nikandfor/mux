package mux

func JoinPath(elems ...string) string {
	size := len(elems) - 1
	for _, e := range elems {
		size += len(e)
	}

	var buf [100]byte
	var b []byte
	if size <= len(buf) {
		b = buf[:]
	} else {
		b = make([]byte, size)
	}

	i := 0

	lastSlash := false

	for _, e := range elems {
		if e == "" {
			continue
		}

		for st, j := 0, 0; j < len(e); st = j + 1 {
			j = index(e, st, '/')
			lastSlash = st == j
			if j == st {
				continue
			}

			switch e[st:j] {
			case ".":
				continue
			case "..":
				for i > 0 && b[i] != '/' {
					i--
				}
				if i > 0 {
					i--
				}

				continue
			}

			b[i] = '/'
			i++

			i += copy(b[i:], e[st:j])
		}
	}

	if i == 0 {
		return "/"
	}

	if lastSlash {
		b[i] = '/'
		i++
	}

	return string(b[:i])
}
