package scheduler

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

//只有master节点才有scheduler

type Scheduler struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
	selectType  string
	queue       queue.ConcurrentQueue
	hashMap     *_map.ConcurrentMap
	mtx         sync.Mutex
}

func NewScheduler(lsConfig *listwatcher.Config) *Scheduler {
	println("[Scheduler] scheduler create")

	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] list watch start fail...")
	}

	scheduler := &Scheduler{
		ls:      ls,
		hashMap: _map.NewConcurrentMap(),
	}
	scheduler.stopChannel = make(chan struct{})
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
			sched.mtx.Lock()
			err := sched.schedulePod(podPtr.(*object.Pod))
			sched.mtx.Unlock()
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

	//已经为pod分配了node
	if pod.Spec.NodeName != "" {
		return
	}

	// check whether scheduled
	fmt.Printf("watch new Pod with name:%s\n", pod.Name)

	sched.queue.Enqueue(pod)
}

func (sched *Scheduler) schedulePod(pod *object.Pod) error {
	fmt.Println("[Scheduler] Begin scheduling")
	if sched.hashMap.Contains(pod.Name) {
		return nil
	}

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

	//填充pod的node部分
	pod.Spec.NodeName = nodeName
	sched.hashMap.Put(pod.Name, nodeName)
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
