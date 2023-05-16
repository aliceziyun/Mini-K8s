package controller

import "context"

//func startDeploymentController(ctx context.Context, controllerContext ControllerContext) (controller.Interface, bool, error) {
//	dc, err := deployment.NewDeploymentController(
//		ctx,
//		controllerContext.InformerFactory.Apps().V1().Deployments(),
//		controllerContext.InformerFactory.Apps().V1().ReplicaSets(),
//		controllerContext.InformerFactory.Core().V1().Pods(),
//		controllerContext.ClientBuilder.ClientOrDie("deployment-controller"),
//	)
//	if err != nil {
//		return nil, true, fmt.Errorf("error creating Deployment controller: %v", err)
//	}
//	go dc.Run(ctx, int(controllerContext.ComponentConfig.DeploymentController.ConcurrentDeploymentSyncs))
//	return nil, true, nil
//}

func startReplicaSetController(ctx context.Context, controllerContext ControllerContext) error {
	//go replicaset.NewReplicaSetController(controllerCtx).Run(ctx)
	return nil
}
