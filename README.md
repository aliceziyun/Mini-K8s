[//]: # (# Mini-K8s)
上海交通大学云操作系统大作业


## Etcd

安装go语言的etcd包应当在终端执行如下命令：
```shell
go get -v github.com/coreos/etcd/clientv3
go mod edit -require=google.golang.org/grpc@v1.26.0
go get -u -x google.golang.org/grpc@v1.26.0 
```
安装完成后执行`go mod tidy`可能会出现报错或警告，但不影响运行，无需担心。
