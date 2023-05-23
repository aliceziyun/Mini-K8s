package scheduler

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/queue"
	"encoding/json"
	"fmt"
	"time"
)

type Scheduler struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
	selectType  string
	queue       queue.ConcurrentQueue
}

func NewScheduler(lsConfig *listwatcher.Config, clientConfig client.Config, selectType string) *Scheduler {
	println("scheduler create")

	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] list watch start fail...")
	}

	scheduler := &Scheduler{
		ls: ls,
	}
	scheduler.stopChannel = make(chan struct{})
	scheduler.selectType = selectType
	return scheduler
}

// Run begins watching and syncing.
func (sched *Scheduler) Run() {
	fmt.Printf("[Scheduler]start running\n")
	go sched.register()
	go sched.worker()
	select {}
}

func (sched *Scheduler) register() {
	err := sched.ls.Watch(_const.POD_CONFIG_PREFIX, sched.watchNewPod, sched.stopChannel)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] listen fail...\n")
	}
}

func (sched *Scheduler) worker() {
	fmt.Printf("[Scheduler] Starting...\n")
	for {
		if !sched.queue.Empty() {
			podPtr := sched.queue.Front()
			sched.queue.Dequeue()
			err := sched.schedulePod(podPtr.(*object.Pod))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (sched *Scheduler) watchNewPod(res etcdstorage.WatchRes) {
	fmt.Printf("[Scheduler] New Pod coming...")
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("watchNewPod bad message pod:%+v\n", pod)
		return
	}

	if pod.Spec.NodeName != "" {
		return
	}

	// check whether scheduled
	fmt.Printf("watch new Config Pod with name:%s\n", pod.Name)

	fmt.Printf("[watchNewPod] new message from watcher...\n")
	sched.queue.Enqueue(pod)
}

func (sched *Scheduler) schedulePod(pod *object.Pod) error {
	fmt.Println("[Scheduler] Begin scheduling")
	nodes, err := sched.getNode()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 选择pod的node
	var nodeName string
	nodeName, err = selectNode(nodes)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("[Scheduler] assign pod to node: %s\n", nodeName)

	pod.Spec.NodeName = nodeName
	err = updatePod(pod)
	return err
}

func (sched *Scheduler) getNode() ([]object.Node, error) {
	raw, err := sched.ls.List(_const.NODE_CONFIG_PREFIX)
	if err != nil {
		return nil, err
	}
	var res []object.Node
	if len(raw) == 0 {
		return res, nil
	}
	for _, rawPair := range raw {
		node := &object.Node{}
		err = json.Unmarshal(rawPair.ValueBytes, node)
		res = append(res, *node)
	}
	return res, nil
}
