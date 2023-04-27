package main

import (
	// kubl "Mini-K8s/pkg/kubelet/dockerClient"
	_ "Mini-K8s/pkg/object"
	"context"
	"fmt"

	//"io"
	//"io/ioutil"
	// _ "github.com/docker/docker/api/types"
	// _ "github.com/docker/docker/api/types/container"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client" // _ "github.com/docker/go-connections/nat"
)

func main() {
	cl, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	fmt.Println(cl.ImageList(context.Background(), types.ImageListOptions{}))

}

// func testNil() {
// 	cli, err2 := kubl.GetNewClient()
// 	if err2 == nil || cli == nil {

// 	}
// 	//cli.ContainerStop(context.Background(), value, container.StopOptions{})
// }

//container.StopOptions{}
