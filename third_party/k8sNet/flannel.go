package k8sNet

//func BootFlannel() error {
//	//先暂停所有的容器
//	err := stopAllContainers()
//	if err != nil {
//		return err
//	}
//	//运行flannel插件
//	go runFlanneld()
//	time.Sleep(10 * time.Second)
//	fmt.Println("run flannel finish")
//	//运行DockerOpt
//	err = runDockerOpt()
//	if err != nil {
//		return err
//	}
//	//修改配置文件
//	err = ModifyDockerServiceConfig()
//	if err != nil {
//		return err
//	}
//	//重启docker 服务
//	err = restartDockerService()
//	if err != nil {
//		return err
//	}
//	return nil
//}
