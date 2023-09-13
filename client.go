package ezops

import (
	"encoding/base64"
	"errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

type KubernetesClientSource struct {
	InCluster      bool
	KubeconfigPath string
	TemporaryDir   string
}

func (s KubernetesClientSource) Build() (client *kubernetes.Clientset, err error) {
	var cfg *rest.Config
	if s.InCluster {
		if cfg, err = rest.InClusterConfig(); err != nil {
			return
		}
	} else {
		if cfg, err = clientcmd.BuildConfigFromFlags("", s.KubeconfigPath); err != nil {
			return
		}
	}
	client, err = kubernetes.NewForConfig(cfg)
	return
}

func (s KubernetesClientSource) CleanUp() {
	if s.TemporaryDir != "" {
		_ = os.RemoveAll(s.TemporaryDir)
	}
}

const (
	tildePrefix = "~" + string(filepath.Separator)
)

func ResolveKubernetesClient(optKubeconfig string) (source KubernetesClientSource, err error) {
	// expand kubeconfig tilde on return
	defer func() {
		if err != nil {
			return
		}
		if source.KubeconfigPath == "" {
			return
		}
		if !strings.HasPrefix(source.KubeconfigPath, tildePrefix) {
			return
		}
		var home string
		if home, err = os.UserHomeDir(); err != nil {
			return
		}
		source.KubeconfigPath = filepath.Join(home, strings.TrimPrefix(source.KubeconfigPath, tildePrefix))
	}()

	var (
		envKubeconfig       = strings.TrimSpace(os.Getenv("KUBECONFIG"))
		envKubeconfigBase64 = strings.TrimSpace(os.Getenv("KUBECONFIG_BASE64"))
	)

	if optKubeconfig != "" {
		source.KubeconfigPath = optKubeconfig
		return
	}

	if envKubeconfig != "" {
		source.KubeconfigPath = envKubeconfig
		return
	}

	if envKubeconfigBase64 != "" {
		var buf []byte
		if buf, err = base64.StdEncoding.DecodeString(envKubeconfigBase64); err != nil {
			return
		}
		if source.TemporaryDir, err = os.MkdirTemp("", "ezops-kubeconfig-*"); err != nil {
			return
		}
		source.KubeconfigPath = filepath.Join(source.TemporaryDir, "kubeconfig")
		if err = os.WriteFile(source.KubeconfigPath, buf, 0600); err != nil {
			return
		}
		return
	}

	if _, err = rest.InClusterConfig(); err == nil {
		source.InCluster = true
		return
	} else {
		if !errors.Is(err, rest.ErrNotInCluster) {
			return
		}
	}

	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return
	}

	source.KubeconfigPath = filepath.Join(home, ".kube", "config")
	return
}
