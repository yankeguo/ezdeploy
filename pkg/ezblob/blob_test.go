package ezblob

import (
	"context"
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"testing"
)

func TestBlob_Load(t *testing.T) {
	dirHome, err := os.UserHomeDir()
	require.NoError(t, err)
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(dirHome, ".kube", "config"))
	require.NoError(t, err)
	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	raw := make([]byte, 755, 755)
	_, err = rand.Read(raw)
	require.NoError(t, err)

	blob := New(Options{
		Client:    client,
		Name:      "ezblob-test",
		Namespace: "default",
		ChunkSize: 128,
	})

	ctx := context.Background()

	err = blob.Save(ctx, raw)
	require.NoError(t, err)

	_, err = rand.Read(raw)
	require.NoError(t, err)

	err = blob.Save(ctx, raw)
	require.NoError(t, err)

	buf, err := blob.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, raw, buf)

	err = blob.Delete(ctx)
	require.NoError(t, err)
}
