package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/yankeguo/ezdeploy"
	"github.com/yankeguo/ezdeploy/pkg/ezkv"
	"github.com/yankeguo/ezdeploy/pkg/ezlog"
	"github.com/yankeguo/ezdeploy/pkg/ezsync"
	"github.com/yankeguo/ezdeploy/pkg/eztmp"
	"github.com/yankeguo/rg"
)

type syncNamespaceOptions struct {
	DB         *ezkv.KV
	Kubeconfig string
	Root       string
	Namespace  string
	Charts     map[string]ezdeploy.Chart
	DryRun     bool
}

func syncNamespace(ctx context.Context, opts syncNamespaceOptions) (err error) {
	defer rg.Guard(&err)

	title := "[" + opts.Namespace + "]"
	log.Println(title, "scanning")

	res := rg.Must(ezdeploy.Load(opts.Root, opts.Namespace, ezdeploy.LoadOptions{Charts: opts.Charts}))

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
	DB         *ezkv.KV
	Resources  []ezdeploy.Resource
	Title      string
	Namespace  string
	Kubeconfig string
	DryRun     bool
}

func syncResources(ctx context.Context, opts syncResourcesOptions) (err error) {
	defer rg.Guard(&err)

	var resources []ezdeploy.Resource

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

	buf := rg.Must(json.Marshal(ezdeploy.NewList(raws)))

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
	DB         *ezkv.KV
	Release    ezdeploy.Release
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

	valuesFile := opts.Release.ValuesFile

	// convert jsonnet file to yaml file
	if strings.HasSuffix(opts.Release.ValuesFile, ezdeploy.SuffixHelmValuesJSONNet) {
		if valuesFile, err = ezdeploy.ConvertJSONNetFileToYAML(valuesFile, opts.Namespace); err != nil {
			return
		}
	}

	args := []string{
		"upgrade", "--install",
		"--namespace", opts.Namespace,
		opts.Release.Name, opts.Release.Chart.Path,
		"-f", valuesFile,
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
	defer eztmp.ClearAll()

	// cli options
	var (
		optDryRun     bool
		optKubeconfig string
	)

	flag.BoolVar(&optDryRun, "dry-run", false, "dry run (server)")
	flag.StringVar(&optKubeconfig, "kubeconfig", "", "path to kubeconfig")
	flag.Parse()

	// context
	ctx := context.Background()

	// kubernetes client source
	cs := rg.Must(ezdeploy.ResolveKubernetesClient(optKubeconfig))
	defer cs.CleanUp()

	if cs.InCluster {
		log.Println("using in-cluster credentials")
	} else {
		log.Println("using kubeconfig:", cs.KubeconfigPath)
	}

	// client
	client := rg.Must(cs.Build())

	// ezkv database
	db := rg.Must(ezkv.Open(ctx, ezkv.Options{
		Client:    client,
		Namespace: "default",
		Name:      "ezdeploydb",
	}))
	defer func() {
		_ = db.Save(ctx)
	}()

	// scan
	result := rg.Must(ezdeploy.Scan("."))

	// sync namespaces
	err = ezsync.DoPara(ctx, result.Namespaces, 5, func(ctx context.Context, namespace string) (err error) {
		return syncNamespace(ctx, syncNamespaceOptions{
			DB:         db,
			Kubeconfig: cs.KubeconfigPath,
			Root:       ".",
			Namespace:  namespace,
			Charts:     result.Charts,
			DryRun:     optDryRun,
		})
	})
}
