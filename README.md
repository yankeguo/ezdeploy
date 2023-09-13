# ezops

[![Go Reference](https://pkg.go.dev/badge/github.com/guoyk93/ezops.svg)](https://pkg.go.dev/github.com/guoyk93/ezops)

A simple GitOps tool, based on one-off run, easy to integrate with existing CI/CD routine

NO DAEMON, NO SERVICE, JUST ONE-OFF EXECUTION

## 中文使用说明

[ezops - 简易 GitOps 工具](https://mp.weixin.qq.com/s/TI2ZozyA5xWltTEwIF46ag)

## Features

* Support `yaml`, `json`, `jsonnet` and `Helm`
* Incremental update

## Installation

You can either build binary from source, or just download pre-built binary.

* Build from source

    ```shell
   git clone https://github.com/guoyk93/ezops.git
   cd ezops
   go build -o ezops ./cmd/ezops
    ```

* Download pre-built binaries

  View <https://github.com/guoyk93/ezops/releases>

## Usage

1. Ensure `kubectl` and `helm` are available in `$PATH`
2. Prepare a **Manifests Directory**, see below
3. Run `ezops`

## Options

* `--dry-run`, run without actually apply any changes.
* `--kubeconfig` or `KUBECONFIG`, specify path to `kubeconfig` file
* `KUBECONFIG_BASE64`, base64 encoded `kubeconfig` content

## Layout of Manifests Directory

* Each top-level directory stands for a **namespace**
* Every manifest file in that directory, will be applied to that **namespace**

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

* Put a `Helm Chart` to top-level directory `_helm`
* Create values file in **namespace** directory, with naming format `[RELEASE_NAME].[CHART_NAME].helm.yaml`

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

`ezops` will create or update a **release** named `main`, using **chart** `_helm/ingress-nginx`, and **values file
** `kube-system/primary.ingress-nginx.helm.yaml`

## Donation

See https://guoyk.xyz/donation

## Credits

GUO YANKE, MIT License
