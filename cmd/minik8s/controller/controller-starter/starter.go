package controller_starter

import (
	"Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/controller/endpoint"
	"Mini-K8s/pkg/controller/replicaset"
	"context"
)

//TODO: 现在如果没有channel控制，就跑不起来

func startReplicaSetController(ctx context.Context, controllerContext controller_context.ControllerContext) error {
	go replicaset.NewReplicaSetController(controllerContext).Run(ctx)
	return nil
}

func StartEndpointController(ctx context.Context, controllerContext controller_context.ControllerContext) error {
	ch := make(chan int)

	go endpoint.NewEndpointController(controllerContext).Run(ctx)

	<-ch

	return nil
}
