# Mini-K8s
上海交通大学云操作系统大作业

## Etcd

安装go语言的etcd包应当在终端执行如下命令：
```shell
go get -v github.com/coreos/etcd/clientv3
go mod edit -require=google.golang.org/grpc@v1.26.0
go get -u -x google.golang.org/grpc@v1.26.0 
```

此外，如果想要运行使用etcd功能，应该在本机安装etcd，并且运行起来，保证其端口与代码中一致（默认值是一致的）。

pod的保存路径：`/registry/pods/{namespace}/{pod-name}`

## Pod Structure

位置：`./pkg/object/object.go`

- Pod
  - apiVersion
  - kind
  - metadata
    - name
    - labels
    - uid
    - namespace
  - spec
    - containers
    - volumes
    - nodeSelector
  - status
    - phase
    - conditions

目前，`Pod` 类型可以导出为 `json` 或者 `yaml`。
