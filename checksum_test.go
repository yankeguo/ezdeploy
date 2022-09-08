package ezops

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestChecksumFile(t *testing.T) {
	s, err := checksumFile(filepath.Join("testdata", "streamfile.txt"))
	require.NoError(t, err)
	require.Equal(t, "3cb95cfbe1035bce8c448fcaf80fe7d9", s)
}

func TestChecksumDir(t *testing.T) {
	h := md5.Sum([]byte("hello\r\nhello\r\n"))
	v, err := checksumDir(filepath.Join("testdata", "checksumdir"))
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(h[:]), v)
}
