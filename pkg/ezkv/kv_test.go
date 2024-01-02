package ezkv

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/yankeguo/ezops/pkg/ezblob"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDatabase(t *testing.T) {
	dirHome, err := os.UserHomeDir()
	require.NoError(t, err)
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(dirHome, ".kube", "config"))
	require.NoError(t, err)
	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	ctx := context.Background()

	db, err := Open(ctx, Options{
		Client:    client,
		Name:      "ezkv-demo",
		Namespace: "default",
	})
	require.NoError(t, err)

	for i := 0; i < 1000; i++ {
		db.Put("hello-"+strconv.Itoa(i), "world-"+strconv.Itoa(i))
	}

	err = db.Save(ctx)
	require.NoError(t, err)

	db, err = Open(ctx, Options{
		Client:    client,
		Name:      "ezkv-demo",
		Namespace: "default",
	})
	require.NoError(t, err)
	require.Equal(t, "world-99", db.Get("hello-99"))

	db.Purge(func(key string, val string) (del bool, stop bool) {
		if strings.HasSuffix(key, "9") {
			del = true
		}
		return
	})

	require.Equal(t, "", db.Get("hello-99"))
	err = db.Save(ctx)
	require.NoError(t, err)

	blob := ezblob.New(ezblob.Options{
		Client:    client,
		Name:      "ezkv-demo",
		Namespace: "default",
	})
	err = blob.Delete(ctx)
	require.NoError(t, err)
}
