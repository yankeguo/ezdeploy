# ezdeploy

一个简单的 Kubernetes 资源部署工具

**无需守护进程，无需启动服务，只需一次性执行**

## 功能

- 支持 `yaml`, `json`, `jsonnet` 和 `Helm`
- 支持增量更新

## 安装

你可以从源码构建二进制文件，或者直接下载预编译的二进制文件。

- 从源码构建

  ```shell
  git clone https://github.com/yankeguo/ezdeploy.git
  cd ezdeploy
  go build -o ezdeploy ./cmd/ezdeploy
  ```

- 下载预编译的二进制文件

  查看 <https://github.com/yankeguo/ezdeploy/releases>

## 用法

1. 确保 `kubectl` 和 `helm` 可以在 `$PATH` 中找到
2. 准备一个 **资源目录**, 参见下文
3. 执行 `ezdeploy`

## 命令参数

- `--dry-run`, 运行但不实际应用任何更改
- `--kubeconfig` 或者 环境变量 `KUBECONFIG`, 指定 `kubeconfig` 文件路径
- `KUBECONFIG_BASE64`, 可以使用此环境变量提供 base64 编码的 `kubeconfig` 文件内容

## 资源文件目录结构

- 每个子目录代表一个**命名空间**
- 子目录中的每个资源文件(`json`,`yaml`,`jsonnet`)，将被应用到该**命名空间**

示例:

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

## Helm 支持

- 下载 Chart 并解压到特殊的子目录 `_helm` 下
- 在 **命名空间** 子目录下，创建 `Values` 文件，命名为 `[RELEASE_NAME].[CHART_NAME].helm.yaml`

示例:

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

`ezdeploy` 会在 `kube-system` 命名空间下，使用 `ingress-nginx` Chart，创建或更新一个名为 `main` 的 Release，`Values` 文件为
`kube-system/primary.ingress-nginx.helm.yaml`

`ezdeploy` 允许使用 `jsonnet` 文件充当 `Values` 文件，只需将文件命名为 `[RELEASE_NAME].[CHART_NAME].helm.jsonnet` 即可。

## 许可证

GUO YANKE, MIT License
