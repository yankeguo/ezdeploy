# ezdeploy

A simple tool to deploy various Kubernetes resources

NO DAEMON, NO SERVICE, JUST ONE-OFF EXECUTION

## Features

- Support `yaml`, `json`, `jsonnet` and `Helm`
- Incremental update

## Installation

You can either build binary from source, or just download pre-built binary.

- Build from source

  ```shell
  git clone https://github.com/yankeguo/ezdeploy.git
  cd ezdeploy
  go build -o ezdeploy ./cmd/ezdeploy
  ```

- Download pre-built binaries

  View <https://github.com/yankeguo/ezdeploy/releases>

## Usage

1. Ensure `kubectl` and `helm` are available in `$PATH`
2. Prepare a **manifests directory**, see below
3. Run `ezdeploy`

## Options

- `--dry-run`, run without actually apply any changes.
- `--kubeconfig` or `KUBECONFIG`, specify path to `kubeconfig` file
- `KUBECONFIG_BASE64`, base64 encoded `kubeconfig` file content

## Layout of Manifests Directory

- Each top-level directory stands for a **namespace**
- Every manifest file in that directory, will be applied to that **namespace**

For example:

```
namespace-a/
  workload-aa.yaml
  workload-ab.jsonnet
  workload-ac.json
namespace-b/
  workload-ba.yaml
  workload-bb.jsonnet
  workload-bc.json
```

## Helm Support

- Put a `Helm Chart` to top-level directory `_helm`
- Create values file in **namespace** directory, with naming format `[RELEASE_NAME].[CHART_NAME].helm.yaml`

For example:

```
_helm/
  ingress-nginx/
    Chart.yaml
    values.yaml
    templates/
      ...
kube-system/
  main.ingress-nginx.helm.yaml
```

`ezdeploy` will create or update a **release** named `main`, using **chart** `_helm/ingress-nginx`, and **values file**
`kube-system/primary.ingress-nginx.helm.yaml`

`ezdeploy` also support values file in `jsonnet`

## Credits

GUO YANKE, MIT License
