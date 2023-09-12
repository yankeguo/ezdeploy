package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/guoyk93/ezops"
	"github.com/guoyk93/ezops/pkg/ezkv"
	"github.com/guoyk93/ezops/pkg/ezlog"
	"github.com/guoyk93/ezops/pkg/ezsync"
	"github.com/guoyk93/rg"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type syncNamespaceOptions struct {
	DB         *ezkv.Database
	Kubeconfig string
	Root       string
	Namespace  string
	Charts     map[string]ezops.Chart
	DryRun     bool
}

func syncNamespace(ctx context.Context, opts syncNamespaceOptions) (err error) {
	defer rg.Guard(&err)

	title := "[" + opts.Namespace + "]"
	log.Println(title, "scanning")

	res := rg.Must(ezops.Load(opts.Root, opts.Namespace, ezops.LoadOptions{Charts: opts.Charts}))

	rg.Must0(syncResources(ctx, syncResourcesOptions{
		DB:         opts.DB,
		Resources:  res.Resources,
		Title:      title,
		Namespace:  opts.Namespace,
		Kubeconfig: opts.Kubeconfig,
		DryRun:     opts.DryRun,
	}))

	rg.Must0(syncResources(ctx, syncResourcesOptions{
		DB:         opts.DB,
		Resources:  res.ResourcesExt,
		Title:      title,
		Namespace:  "",
		Kubeconfig: opts.Kubeconfig,
		DryRun:     opts.DryRun,
	}))

	for _, release := range res.Releases {
		rg.Must0(syncRelease(ctx, syncReleaseOptions{
			DB:         opts.DB,
			Release:    release,
			Title:      title + " [Helm:" + release.Name + "]",
			Namespace:  opts.Namespace,
			Kubeconfig: opts.Kubeconfig,
			DryRun:     opts.DryRun,
		}))
	}

	return
}

type syncResourcesOptions struct {
	DB         *ezkv.Database
	Resources  []ezops.Resource
	Title      string
	Namespace  string
	Kubeconfig string
	DryRun     bool
}

func syncResources(ctx context.Context, opts syncResourcesOptions) (err error) {
	defer rg.Guard(&err)

	var resources []ezops.Resource

	for _, res := range opts.Resources {
		if opts.DB.Get(res.ID) == res.Checksum {
			continue
		}
		resources = append(resources, res)
	}

	if len(resources) == 0 {
		return
	}

	var raws []json.RawMessage
	for _, res := range resources {
		raws = append(raws, res.Raw)
	}

	buf := rg.Must(json.Marshal(ezops.NewList(raws)))

	args := []string{"apply", "-f", "-"}

	if opts.Kubeconfig != "" {
		args = append([]string{"--kubeconfig", opts.Kubeconfig}, args...)
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.DryRun {
		args = append(args, "--dry-run=server")
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = ezlog.NewLogWriter(log.Default(), opts.Title)
	cmd.Stderr = ezlog.NewLogWriter(log.Default(), opts.Title)
	rg.Must0(cmd.Run())

	if !opts.DryRun {
		for _, res := range resources {
			opts.DB.Put(res.ID, res.Checksum)
		}
	}

	if opts.DryRun {
		log.Println(opts.Title, "resources synced (dry run)")
	} else {
		log.Println(opts.Title, "resources synced")
	}

	return
}

type syncReleaseOptions struct {
	DB         *ezkv.Database
	Release    ezops.Release
	Title      string
	Namespace  string
	Kubeconfig string
	DryRun     bool
}

func syncRelease(ctx context.Context, opts syncReleaseOptions) (err error) {
	defer rg.Guard(&err)

	if opts.DB.Get(opts.Release.ID) == opts.Release.Checksum {
		return
	}

	args := []string{
		"upgrade", "--install",
		"--namespace", opts.Namespace,
		opts.Release.Name, opts.Release.Chart.Path,
		"-f", opts.Release.ValuesFile,
	}

	if opts.Kubeconfig != "" {
		args = append([]string{"--kubeconfig", opts.Kubeconfig}, args...)
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = ezlog.NewLogWriter(log.Default(), opts.Title)
	cmd.Stderr = ezlog.NewLogWriter(log.Default(), opts.Title)
	rg.Must0(cmd.Run())

	if !opts.DryRun {
		opts.DB.Put(opts.Release.ID, opts.Release.Checksum)
	}

	if opts.DryRun {
		log.Println(opts.Title, "release synced (dry run)")
	} else {
		log.Println(opts.Title, "release synced")
	}

	return
}

func resolveKubernetesClient(envKubeconfig, optKubeconfig string) (client *kubernetes.Clientset, flagKubeconfig string, err error) {
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return
	}
	if strings.HasPrefix(envKubeconfig, "~/") {
		envKubeconfig = filepath.Join(home, envKubeconfig[2:])
	}
	if strings.HasPrefix(optKubeconfig, "~/") {
		optKubeconfig = filepath.Join(home, optKubeconfig[2:])
	}

	var opts *rest.Config
	if optKubeconfig != "" {
		if opts, err = clientcmd.BuildConfigFromFlags("", optKubeconfig); err != nil {
			return
		}
		flagKubeconfig = optKubeconfig
	} else if envKubeconfig != "" {
		if opts, err = clientcmd.BuildConfigFromFlags("", envKubeconfig); err != nil {
			return
		}
		flagKubeconfig = envKubeconfig
	} else {
		if opts, err = rest.InClusterConfig(); err != nil {
			if errors.Is(err, rest.ErrNotInCluster) {
				path := filepath.Join(home, ".kube", "config")
				if opts, err = clientcmd.BuildConfigFromFlags("", path); err != nil {
					return
				}
				flagKubeconfig = path
			} else {
				return
			}
		}
	}

	client, err = kubernetes.NewForConfig(opts)
	return
}

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()
	defer rg.Guard(&err)

	// cli options
	var (
		optDryRun     bool
		optKubeconfig string

		envKubeconfig = strings.TrimSpace(os.Getenv("KUBECONFIG"))
	)

	flag.BoolVar(&optDryRun, "dry-run", false, "dry run (server)")
	flag.StringVar(&optKubeconfig, "kubeconfig", "", "path to kubeconfig")
	flag.Parse()

	// context
	ctx := context.Background()

	// kubernetes client
	client, flagKubeconfig := rg.Must2(resolveKubernetesClient(envKubeconfig, optKubeconfig))

	// ezkv database
	db := rg.Must(ezkv.Open(ctx, ezkv.Options{
		Client:    client,
		Namespace: "default",
		Name:      "ezopsdb",
	}))
	defer func() {
		_ = db.Save(ctx)
	}()

	// scan
	result := rg.Must(ezops.Scan("."))

	// sync namespaces
	err = ezsync.DoPara(ctx, result.Namespaces, 5, func(ctx context.Context, namespace string) (err error) {
		return syncNamespace(ctx, syncNamespaceOptions{
			DB:         db,
			Kubeconfig: flagKubeconfig,
			Root:       ".",
			Namespace:  namespace,
			Charts:     result.Charts,
			DryRun:     optDryRun,
		})
	})
}
