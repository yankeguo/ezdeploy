package ezblob

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomRevision(t *testing.T) {
	revision := randomRevision()
	require.Len(t, revision, 7)
}

func TestSplitBytes(t *testing.T) {
	v := splitBytes([]byte("hello,world"), 4)
	require.Equal(t, [][]byte{[]byte("hell"), []byte("o,wo"), []byte("rld")}, v)
	v = splitBytes([]byte("hello,wo"), 4)
	require.Equal(t, [][]byte{[]byte("hell"), []byte("o,wo")}, v)
	v = splitBytes([]byte{}, 4)
	require.Nil(t, v)
}
