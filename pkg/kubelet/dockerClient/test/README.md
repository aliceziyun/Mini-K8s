# 三个功能
1. 通过pod的yaml配置创建pod内的container
2. 返回刚创建的container的所有信息
3. 根据container的id删除一个container

在 `Mini-K8s/pkg/kubelet/dockerClient/test/testDockerClient` 中。
会用到 `Mini-K8s/pkg/kubelet/dockerClient/test/getContainersByYaml` 以及 `Mini-K8s/pkg/object` 。
其中 `createContainersOfPod` 函数实现功能 1 ，函数中会对刚创建的containers，根据containerId 分别调用函数 `getContainerInfo` 来获取容器信息；也可以使用函数`getAllContainers` 返回**所有**容器（信息），包括以前创建的容器。
函数 `deleteContainerById` 和 `deleteContainersByIds` 可以根据containerId删除容器。

上述几种函数在文件的前几个函数。

对于以上三个功能的测试在`main`函数中，简单测试过，应该是可以的。main函数在文件的最后一个函数。
````
func main() {
	containers := parseYaml.GetContainersByFile(yamlPath)

	createContainersOfPod(containers)

	fmt.Println("======delete test=====")
	tmp, err := getAllContainers()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("----------------\nthis is the container to be deleted: ")
	fmt.Printf("image=%s,id=%s\n", tmp[1].Image, tmp[1].ID)
	fmt.Println("----------------")
	deleteContainerById(tmp[1].ID)
	getAllContainers()
	fmt.Println("======delete test=====")
}
````

另外，对yaml文件的简单解析在 `Mini-K8s/pkg/kubelet/dockerClient/test/getContainersByYaml` 目录下的 `getContainersByYaml.go` 中，只对简单yaml进行解析（没有volumes之类的）