package ezblob

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestBlob(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	raw := make([]byte, 755)
	_, err = rand.Read(raw)
	require.NoError(t, err)

	blob, err := New(Options{
		Client:    client,
		Name:      "ezblob-test",
		Namespace: "default",
		ChunkSize: 128,
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = blob.Save(ctx, raw)
	require.NoError(t, err)

	buf, err := blob.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, raw, buf)

	_, err = rand.Read(raw)
	require.NoError(t, err)

	err = blob.Save(ctx, raw)
	require.NoError(t, err)

	buf, err = blob.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, raw, buf)

	err = blob.Delete(ctx)
	require.NoError(t, err)
}
