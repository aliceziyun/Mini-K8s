package job

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"context"
	"fmt"
	"time"
)

type JobController struct {
	ls           *listwatcher.ListWatcher
	jobMap       *_map.ConcurrentMapTrait[string, object.GPUJob]
	jobStatusMap *_map.ConcurrentMapTrait[string, object.JobStatus]
	stopChannel  chan struct{}
	allocator    *object.AccountAllocator
}

func NewJobController(controllerCtx controller_context.ControllerContext) *JobController {
	jc := &JobController{
		ls:           controllerCtx.Ls,
		stopChannel:  make(chan struct{}),
		jobMap:       _map.NewConcurrentMapTrait[string, object.GPUJob](),
		jobStatusMap: _map.NewConcurrentMapTrait[string, object.JobStatus](),
		allocator:    object.NewAccountAllocator(),
	}

	return jc
}

func (jc *JobController) Run(ctx context.Context) {
	fmt.Println("[Job Controller] start run...")
	jc.register()
	<-ctx.Done()
	close(jc.stopChannel)
}

func (jc *JobController) register() {
	// register job handler
	go func() {
		for {
			err := jc.ls.Watch(_const.JOB_CONFIG, jc.handleJob, jc.stopChannel)
			if err != nil {
				fmt.Println("[Job Controller] list watch RS handler init fail...")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (jc *JobController) handleJob(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}
	
}
