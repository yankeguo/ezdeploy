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

    View https://github.com/guoyk93/ezops/releases

## Usage

1. Make sure you have valid `kubectl` and `helm` command in your `$PATH`
2. Prepare your **resource directory**, see below
3. Just run `ezops` command

## Options

* `--dry-run`, run without actually apply any changes.
* `--kubeconfig`, specify path to `kubeconfig` file

## Layout of Resource Directory

* Each top-level directory stands for a **Namespace**
* Each resource file in **Namespace** directory, will be applied to that **Namespace**

For example:

```
myapp/
  service-a.yaml
  service-b.jsonnet
  service-c.json
```

`ezops` will read `service-a.yaml`, `service-b.jsonnet`, `service-c.json` files, and apply resources to namespace `myapp`

## Helm Support

* Put your `Helm` `Chart` to a special top-level directory `_helm`
* Create values file in **Namespace** directory, with naming format `[RELEASE_NAME].[CHART_NAME].helm.yaml`

For exmaple:

```
_helm/
  ingress-nginx/
    Chart.yaml
    values.yaml
    templates/
      ...
kube-system/
  primary.ingress-nginx.helm.yaml
```

`ezops` will create or update a **Release** named `primary` in **Namespace** `kube-system`,

using **Chart** at `_helm/ingress-nginx`, and **Values File** at `kube-system/primary.ingress-nginx.helm.yaml`

## Upstream

https://git.guoyk.net/go-guoyk/gops

Due to various reasons, codebase is detached from upstream.

## Donation

![oss-donation-wx](https://www.guoyk.net/oss-donation-wx.png)

## Credits

Guo Y.K., MIT License
