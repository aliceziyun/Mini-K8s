package scheduler

import (
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listener"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/queue"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Scheduler struct {
	ls          *listener.Listener
	stopChannel <-chan struct{}
	selectType  string
	queue       queue.ConcurrentQueue
	Client      client.RESTClient
}

func NewScheduler(lsConfig *listener.Config, clientConfig client.Config, selectType string) *Scheduler {
	println("scheduler create")

	ls, err := listener.NewListener(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] list watch start fail...")
	}

	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}

	scheduler := &Scheduler{
		ls:     ls,
		Client: restClient,
	}
	scheduler.stopChannel = make(chan struct{})
	scheduler.selectType = selectType
	return scheduler
}

// Run begins watching and syncing.
func (sched *Scheduler) Run(ctx context.Context) {
	fmt.Printf("[Scheduler]start running\n")
	go sched.register()
	go sched.worker(ctx)
	select {}
}

func (sched *Scheduler) register() {
	podConfig := "/testwatch"
	err := sched.ls.Watch(podConfig, sched.watchNewPod, sched.stopChannel)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] listen fail...\n")
	}
}

func (sched *Scheduler) worker(ctx context.Context) {
	fmt.Printf("[worker] Starting...\n")
	for {
		if !sched.queue.Empty() {
			podPtr := sched.queue.Front()
			sched.queue.Dequeue()
			//sched.schedulePod(ctx, podPtr.(*object.Pod))
			createPod(ctx, podPtr.(*object.Pod))
		} else {
			time.Sleep(time.Second)
		}
	}
}

// watch the change of new pods
func (sched *Scheduler) watchNewPod(res etcdstorage.WatchRes) {
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("watchNewPod bad message pod:%+v\n", pod)
		return
	}

	//if pod.Spec.NodeName != "" {
	//	return
	//}

	// check whether scheduled
	fmt.Printf("watch new Config Pod with name:%s\n", pod.Name)

	fmt.Printf("[watchNewPod] new message from watcher...\n")
	sched.queue.Enqueue(pod)
}

// 根据配置创建pod，测试用
func createPod(ctx context.Context, pod *object.Pod) {

}
