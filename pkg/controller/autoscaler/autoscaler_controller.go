package autoscaler

import (
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/listwatcher"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"context"
	"fmt"
)

type AutoScaleController struct {
	ls           *listwatcher.ListWatcher
	stopChannel  <-chan struct{}
	queue        queue.ConcurrentQueue
	hashMap      *_map.ConcurrentMap
	metricClient client.MetricClient
}

func NewAutoScaleController(controllerContext controller_context.ControllerContext) *AutoScaleController {
	hash := _map.NewConcurrentMap()
	mClient := client.MetricClient{Base: "localhost:8080"}
	asc := &AutoScaleController{
		ls:           controllerContext.Ls,
		hashMap:      hash,
		metricClient: mClient,
	}
	return asc
}

func (asc *AutoScaleController) Run(ctx context.Context) {
	fmt.Println("[AutoScale Controller] start run ...")
}
