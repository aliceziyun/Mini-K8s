package main

import (
	controllerConfig "Mini-K8s/cmd/minik8s/controller/config"
	"Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/cmd/minik8s/controller/controller-starter"
	"Mini-K8s/pkg/listwatcher"
	"context"
	"fmt"
)

func main() {
	// only for test
	fmt.Println("[Controller] test start")
	controllerCtx := getControllerContext()
	//err := controller_starter.StartEndpointController(context.TODO(), *controllerCtx)
	err := controller_starter.StartReplicaSetController(context.TODO(), *controllerCtx)
	if err != nil {
		fmt.Println("[Controller] start fail")
		return
	}
}

func getControllerContext() *controller_context.ControllerContext {
	ls, err := listwatcher.NewListWatcher(listwatcher.DefaultConfig())
	if err != nil {
		return nil
	}
	option := controllerConfig.NewKubeControllerManagerOptions()
	c := option.Config()
	controllerContext := &controller_context.ControllerContext{
		Ls:     ls,
		Config: c.Complete(),
	}
	return controllerContext
}
