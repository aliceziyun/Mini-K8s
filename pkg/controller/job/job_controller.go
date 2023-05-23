package job

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"context"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"path"
	"time"
)

type JobController struct {
	ls          *listwatcher.ListWatcher
	stopChannel chan struct{}
}

func NewJobController(controllerCtx controller_context.ControllerContext) *JobController {
	jc := &JobController{
		ls:          controllerCtx.Ls,
		stopChannel: make(chan struct{}),
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
			err := jc.ls.Watch(_const.JOB_CONFIG_PREFIX, jc.handleJob, jc.stopChannel)
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
	//account := getAccount(job.Spec.SlurmConfig.Partition)

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
	fmt.Printf("[Job Controller] new container %s \n", container.Name)
	//commands := []string{
	//	"sh",
	//	"test.sh",
	//	account.GetUsername(),
	//	account.GetPassword(),
	//	account.GetHost(),
	//	"/home/job",
	//	path.Join(account.GetRemoteBasePath(), "job"+"test"),
	//}
	commands := []string{
		"/bin/sh",
		"-c",
		"while true; do echo hello world; sleep 1; done",
	}
	container.Command = commands
	//container.Command = nil
	container.Args = nil
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
			Path: path.Join(_const.SHARED_DATA_DIR, "job-"+job.Metadata.Uid),
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
