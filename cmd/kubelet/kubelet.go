//有些暂时用来参考
package main

import (
	"Mini-K8s/pkg/kubelet/dockerClient"
	"Mini-K8s/pkg/object"
)

const MasterIp = "10.119.11.108"

//func parseConfigFile(path string) *object.Node {
//	data, err := ioutil.ReadFile(path)
//	if err != nil {
//		fmt.Printf("ReadFile in %s fail, use default config", os.Args[1])
//		return nil
//	}
//	node := &object.Node{}
//	err = yaml.Unmarshal([]byte(data), node)
//	if err != nil {
//		fmt.Printf("file in %s unmarshal fail, use default config", path)
//		return nil
//	}
//	return node
//}

//func main() {
//	//var node *object.Node
//	node := &object.Node{}
//	masterIp := MasterIp
//	if len(os.Args) != 1 {
//		//参数应该为yaml文件路径,进行解析
//		//node = parseConfigFile(os.Args[1])
//
//		if node != nil {
//			masterIp = node.MasterIp
//		}
//	}
//	//clientConfig := client.Config{Host: masterIp + ":8080"}
//	clientConfig := masterIp + ":8080"
//	kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(masterIp), clientConfig, node)
//	kube.Run()
//	//fmt.Printf("kube run emd...\n")
//	select {}
//}
func main() {
	//var node *object.Node
	//node := &object.Node{}
	//masterIp := MasterIp
	// if len(os.Args) != 1 {
	// 	path := os.Args[1]
	// 	//参数应该为yaml文件路径,进行解析
	// 	//node = parseConfigFile(os.Args[1])

	// 	if err != nil {
	// 		fmt.Printf("file in %s unmarshal fail, use default config", path)
	// 		//node = nil
	// 	}

	// }
	containers := []object.Container{
		{
			Name:    "container1",
			Image:   "img1",
			Ports:   nil,
			Env:     nil,
			Command: nil,
			Args:    nil,
		},
		{
			Name:    "container2",
			Image:   "img2",
			Ports:   nil,
			Env:     nil,
			Command: nil,
			Args:    nil,
		},
		{
			Name:    "container2",
			Image:   "img2",
			Ports:   nil,
			Env:     nil,
			Command: nil,
			Args:    nil,
		},
	}
	dockerClient.Main(containers)
	//clientConfig := client.Config{Host: masterIp + ":8080"}
	//clientConfig := masterIp + ":8080"
	//kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(masterIp), clientConfig, node)
	//kube.Run()
	//fmt.Printf("kube run emd...\n")
	select {}
}
