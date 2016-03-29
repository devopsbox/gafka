package main

import (
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestTagMessage(t *testing.T) {
	msg := []byte("hello world")
	tags := map[string]string{
		"a": "b",
		"c": "d",
	}

	m := TagMessage(tags, msg)
	t.Logf("%s", string(m.Body))
	assert.Equal(t, `{"a":"b","c":"d"}hello world`, string(m.Body[1:]))
	assert.Equal(t, byte(0), m.Body[0])

	assert.Equal(t, true, isTaggedMessage(m.Body))

	// untag
	rawMsg, err := UntagMessage(m.Body, &tags)
	assert.Equal(t, nil, err)
	assert.Equal(t, string(msg), string(rawMsg))
	assert.Equal(t, "b", tags["a"])
	assert.Equal(t, "d", tags["c"])

	m.Free()
}

func BenchmarkTagMessage(b *testing.B) {
	msg := []byte(strings.Repeat("X", 1250))
	tags := map[string]string{
		"a": "b",
		"c": "d",
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := TagMessage(tags, msg)
		m.Free()
	}
	b.SetBytes(1250)
}
