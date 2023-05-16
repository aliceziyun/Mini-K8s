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

每次运行前记得启动etcd！！！

## rabbitMQ
安装及权限配置
用户名和密码看 `message.config`
https://www.0758q.com/zixun/1913.html

## 创建POD
开三个命令行
```shell
   go run kubectl.go
   go run kubelet.go
   go run api-server.go
```
然后在kubectl的窗口中输入:
```shell
   kubectl apply
```
稍等片刻便能看到build文件夹下的文件被读入并创建pod