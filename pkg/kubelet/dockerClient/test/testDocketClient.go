package main

import (
	kubl "Mini-K8s/pkg/kubelet/dockerClient"
)

func testNil() {
	cli, err2 := kubl.GetNewClient()
	if err2 == nil || cli == nil {

	}
	//cli.ContainerStop(context.Background(), value, container.StopOptions{})
}

//container.StopOptions{}
