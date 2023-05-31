[//]: # (# Mini-K8s)
上海交通大学云操作系统大作业

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
   kubectl apply -f filename
```
稍等片刻便能看到build文件夹下的文件被读入并创建pod
