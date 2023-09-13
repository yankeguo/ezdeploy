package ezops

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveKubernetesClient(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	src, err := ResolveKubernetesClient("")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".kube", "config"), src.KubeconfigPath)
	require.False(t, src.InCluster)
	require.Empty(t, src.TemporaryDir)

	err = os.Setenv("KUBECONFIG_BASE64", base64.StdEncoding.EncodeToString([]byte("hello")))
	require.NoError(t, err)

	src, err = ResolveKubernetesClient("")
	require.NoError(t, err)
	require.False(t, src.InCluster)
	require.DirExists(t, src.TemporaryDir)
	buf, err := os.ReadFile(src.KubeconfigPath)
	require.NoError(t, err)
	require.Equal(t, "hello", string(buf))
	src.CleanUp()
	require.NoDirExists(t, src.TemporaryDir)

	err = os.Setenv("KUBECONFIG", "~/.kube/config333")
	require.NoError(t, err)

	src, err = ResolveKubernetesClient("~/.kube/config222")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".kube", "config222"), src.KubeconfigPath)
	require.Empty(t, src.TemporaryDir)

	src, err = ResolveKubernetesClient("")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".kube", "config333"), src.KubeconfigPath)
	require.Empty(t, src.TemporaryDir)
}
