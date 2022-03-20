package mux

import (
	"testing"

	"github.com/nikandfor/assert"
)

func TestJoinPath(t *testing.T) {
	assert.Equal(t, "/", JoinPath("/"))
	assert.Equal(t, "/", JoinPath("/", "/"))
	assert.Equal(t, "/", JoinPath("/", ""))
	assert.Equal(t, "/", JoinPath("/", "/", "", "/"))

	assert.Equal(t, "/qwe", JoinPath("/", "/qwe"))
	assert.Equal(t, "/qwe/asd", JoinPath("/", "qwe", "asd"))
	assert.Equal(t, "/qwe/asd", JoinPath("/", "/qwe", "/asd"))
	assert.Equal(t, "/qwe/asd", JoinPath("/", "qwe/", "/asd"))
	assert.Equal(t, "/qwe/asd", JoinPath("/", "qwe", "/asd"))

	assert.Equal(t, "/qwe/asd/", JoinPath("/", "qwe/", "asd/"))
	assert.Equal(t, "/qwe/asd/", JoinPath("/", "qwe/", "asd", "/"))

	assert.Equal(t, "/", JoinPath("/", "..", ".."))
	assert.Equal(t, "/", JoinPath("..", ".."))
	assert.Equal(t, "/a", JoinPath("..", "..", "a"))

	assert.Equal(t, "/qwe/asd/", JoinPath("/", "qwe/", ".", "asd", "/"))
	assert.Equal(t, "/asd/", JoinPath("/", "qwe/", "..", "asd", "/"))
	assert.Equal(t, "/asd/", JoinPath("/", "qwe/", "..", "..", "asd/"))
}

func BenchmarkJoinPath(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = JoinPath([]string{"/", "v0/", "some/path", "another/path", "and/a/bit/more/"}...)
	}
}
