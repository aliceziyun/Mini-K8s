// 有些暂时用来参考
package main

import (
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/kubelet"
	"Mini-K8s/pkg/listwatcher"
)

//const MasterIp = "10.119.11.108"

//	func main() {
//		//var node *object.Node
//		node := &object.Node{}
//		masterIp := MasterIp
//		if len(os.Args) != 1 {
//			//参数应该为yaml文件路径,进行解析
//			//node = parseConfigFile(os.Args[1])
//
//			if node != nil {
//				masterIp = node.MasterIp
//			}
//		}
//		//clientConfig := client.Config{Host: masterIp + ":8080"}
//		clientConfig := masterIp + ":8080"
//		kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(masterIp), clientConfig, node)
//		kube.Run()
//		//fmt.Printf("kube run emd...\n")
//		select {}
//	}

func main() {
	clientConfig := client.Config{Host: "192.168.1.6" + ":8080"}
	kube := kubelet.NewKubelet(listwatcher.DefaultConfig(), clientConfig)
	kube.Run()
}
