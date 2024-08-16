package ezblob

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomRevision(t *testing.T) {
	revision, err := randomRevision()
	require.NoError(t, err)
	require.Len(t, revision, 7)
}

func TestChunkify(t *testing.T) {
	v := chunkify([]byte("hello,world"), 4)
	require.Equal(t, [][]byte{[]byte("hell"), []byte("o,wo"), []byte("rld")}, v)
	v = chunkify([]byte("hello,wo"), 4)
	require.Equal(t, [][]byte{[]byte("hell"), []byte("o,wo")}, v)
	v = chunkify([]byte{}, 4)
	require.Nil(t, v)
}
