package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/guoyk93/ezops"
	"github.com/guoyk93/ezops/pkg/ezkv"
	"github.com/guoyk93/grace"
	"github.com/guoyk93/grace/gracelog"
	"github.com/guoyk93/grace/gracemain"
	"github.com/guoyk93/grace/gracesync"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	defer grace.Guard(&err)

	title := "[" + opts.Namespace + "]"
	log.Println(title, "scanning")

	res := grace.Must(ezops.Load(opts.Root, opts.Namespace, ezops.LoadOptions{Charts: opts.Charts}))

	grace.Must0(syncResources(ctx, syncResourcesOptions{
		DB:         opts.DB,
		Resources:  res.Resources,
		Title:      title,
		Namespace:  opts.Namespace,
		Kubeconfig: opts.Kubeconfig,
		DryRun:     opts.DryRun,
	}))

	grace.Must0(syncResources(ctx, syncResourcesOptions{
		DB:         opts.DB,
		Resources:  res.ResourcesExt,
		Title:      title,
		Namespace:  "",
		Kubeconfig: opts.Kubeconfig,
		DryRun:     opts.DryRun,
	}))

	for _, release := range res.Releases {
		grace.Must0(syncRelease(ctx, syncReleaseOptions{
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
	defer grace.Guard(&err)

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

	buf := grace.Must(json.Marshal(ezops.NewList(raws)))

	args := []string{"--kubeconfig", opts.Kubeconfig, "apply", "-f", "-"}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.DryRun {
		args = append(args, "--dry-run=server")
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = gracelog.NewLogWriter(log.Default(), opts.Title)
	cmd.Stderr = gracelog.NewLogWriter(log.Default(), opts.Title)
	grace.Must0(cmd.Run())

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
	defer grace.Guard(&err)

	if opts.DB.Get(opts.Release.ID) == opts.Release.Checksum {
		return
	}

	args := []string{
		"--kubeconfig", opts.Kubeconfig,
		"upgrade", "--install",
		"--namespace", opts.Namespace,
		opts.Release.Name, opts.Release.Chart.Path,
		"-f", opts.Release.ValuesFile,
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = gracelog.NewLogWriter(log.Default(), opts.Title)
	cmd.Stderr = gracelog.NewLogWriter(log.Default(), opts.Title)
	grace.Must0(cmd.Run())

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

func main() {
	var err error
	defer gracemain.Exit(&err)
	defer grace.Guard(&err)

	// determine user home
	dirHome := grace.Must(os.UserHomeDir())

	// cli options
	var (
		optDryRun     bool
		optKubeconfig string
	)

	flag.BoolVar(&optDryRun, "dry-run", false, "dry run (server)")
	flag.StringVar(&optKubeconfig, "kubeconfig", filepath.Join(dirHome, ".kube", "config"), "path to kubeconfig")
	flag.Parse()

	// context
	ctx := context.Background()

	// kubernetes client
	client := grace.Must(kubernetes.NewForConfig(grace.Must(clientcmd.BuildConfigFromFlags("", optKubeconfig))))

	// ezkv database
	db := grace.Must(ezkv.Open(ctx, ezkv.Options{
		Client:    client,
		Namespace: "default",
		Name:      "ezopsdb",
	}))
	defer func() {
		_ = db.Save(ctx)
	}()

	// scan
	result := grace.Must(ezops.Scan("."))

	// sync namespaces
	err = gracesync.DoPara(ctx, result.Namespaces, 5, func(ctx context.Context, namespace string) (err error) {
		return syncNamespace(ctx, syncNamespaceOptions{
			DB:         db,
			Kubeconfig: optKubeconfig,
			Root:       ".",
			Namespace:  namespace,
			Charts:     result.Charts,
			DryRun:     optDryRun,
		})
	})
}
