package job

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"context"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"path"
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
	job := object.GPUJob{}
	err := json.Unmarshal(res.ValueBytes, &job)
	if err != nil {
		fmt.Println(err)
		return
	}
	//account, err := jc.allocator.Allocate(job.Spec.SlurmConfig.Partition)

	// 对Pod进行初始化
	pod := object.Pod{}
	pod.Metadata.Name = fmt.Sprintf("Job-%s-Pod", job.Metadata.Uid)
	uuid, err := uuid2.NewUUID()
	if err != nil {
		fmt.Println(err)
		return
	}
	pod.Metadata.Uid = uuid.String()
	pod.Name = fmt.Sprintf("Job-%s-Pod", job.Metadata.Uid)
	pod.Kind = object.POD

	container := job.Spec.App.AppSpec.Container
	args := []string{
		"/root/remote_runner",
		account.GetUsername(),
		account.GetPassword(),
		account.GetHost(),
		"/home/job",
		path.Join(account.GetRemoteBasePath(), path.Base(res.Key)),
	}
	container.Args = args
	container.Command = nil
	volumeMounts := []object.VolumeMount{
		{
			Name:      "gpuPath",
			MountPath: "/home/job",
		},
	}
	container.VolumeMounts = volumeMounts
	container.Ports = []object.ContainerPort{
		{ContainerPort: "9999"},
	}
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	volumes := []object.Volume{
		{
			Name: "gpuPath",
			Type: "hostPath",
			Path: path.Join(_const.SHARED_DATA_DIR, path.Base(res.Key)),
		},
	}
	pod.Spec.Volumes = volumes

	go func() {
		fmt.Println("[Job Controller] add new GPU job")
		err = addJobPod(&pod)
		if err != nil {
			fmt.Println(err)
			return
		}

		//TODO: 这里可能还需要维护一个job到pod的映射
	}()
}
