package ezops

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	res, err := Scan(filepath.Join("testdata", "root"))
	require.NoError(t, err)
	chart, ok := res.Charts["demo-chart"]
	require.True(t, ok)
	require.Equal(t, "demo-chart", chart.Name)
	require.Equal(t, []string{"default"}, res.Namespaces)
}
