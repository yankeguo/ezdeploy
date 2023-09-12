package ezlog

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestLogWriter(t *testing.T) {
	out := &bytes.Buffer{}
	l := log.New(out, "aaa", 0)
	w := NewLogWriter(l, "bbb")
	_, err := w.Write([]byte("hello,world\nbbb"))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	require.Equal(t, "aaabbb hello,world\naaabbb bbb\n", out.String())
}
