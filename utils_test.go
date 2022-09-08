package ezops

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestReadDirNames(t *testing.T) {
	names, err := readDirNames("testdata")
	require.NoError(t, err)
	require.Equal(t, []string{"checksumdir", "root", "subdir"}, names)
}

func TestStreamFile(t *testing.T) {
	buf := &bytes.Buffer{}
	err := streamFile(buf, filepath.Join("testdata", "streamfile.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello,world", buf.String())
}
