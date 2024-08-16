package ezkv

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yankeguo/ezdeploy/pkg/ezblob"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestKV(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
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

	blob, err := ezblob.New(ezblob.Options{
		Client:    client,
		Name:      "ezkv-demo",
		Namespace: "default",
	})
	require.NoError(t, err)
	err = blob.Delete(ctx)
	require.NoError(t, err)
}
