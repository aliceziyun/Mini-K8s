package controller_starter

import (
	"Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/controller/autoscaler"
	"Mini-K8s/pkg/controller/endpoint"
	"Mini-K8s/pkg/controller/job"
	"Mini-K8s/pkg/controller/replicaset"
	"context"
)

//TODO: 现在如果没有channel控制，就跑不起来

func StartReplicaSetController(ctx context.Context, controllerContext controller_context.ControllerContext) {
	ch := make(chan int)

	go replicaset.NewReplicaSetController(controllerContext).Run(ctx)

	<-ch

	return
}

func StartEndpointController(ctx context.Context, controllerContext controller_context.ControllerContext) {
	ch := make(chan int)

	go endpoint.NewEndpointController(controllerContext).Run(ctx)

	<-ch

	return
}

func StartAutoScaleController(ctx context.Context, controllerContext controller_context.ControllerContext) {
	ch := make(chan int)

	go autoscaler.NewAutoScaleController(controllerContext).Run(ctx)

	<-ch

	return
}

func StartJobController(ctx context.Context, controllerContext controller_context.ControllerContext) {
	ch := make(chan int)

	go job.NewJobController(controllerContext).Run(ctx)

	<-ch

	return
}
