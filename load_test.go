package ezops

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	res, err := Scan(filepath.Join("testdata", "root"))
	require.NoError(t, err)

	res1, err := Load(filepath.Join("testdata", "root"), res.Namespaces[0], LoadOptions{
		Charts: res.Charts,
	})
	require.NoError(t, err)
	_ = res1
}
