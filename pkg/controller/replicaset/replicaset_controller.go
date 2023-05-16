package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/cmd/minik8s/controller/controller"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/controller/replicaset/RSConfig"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"context"
	"encoding/json"
	"fmt"
)

type ReplicaSetController struct {
	ls          *listwatcher.ListWatcher
	config      *RSConfig.RSConfig
	stopChannel <-chan struct{}
	client      client.RESTClient
	queue       queue.ConcurrentQueue
	hashMap     _map.ConcurrentMap
}

func NewReplicaSetController(controllerContext controller.ControllerContext) *ReplicaSetController {
	rsConfig := RSConfig.NewRSConfig()
	restClient := client.RESTClient{
		Base: "http://" + controllerContext.MasterIP + ":" + controllerContext.HttpServerPort,
	}
	rsc := &ReplicaSetController{
		ls:     controllerContext.Ls,
		config: rsConfig,
		client: restClient,
	}
	return rsc
}

func (rsc *ReplicaSetController) Run(ctx context.Context) {
	fmt.Println("[ReplicaSet] start run ...")
	//rsUpdate := rsc.config.GetUpdates()

	// 接受到新消息，开始处理
	if !rsc.queue.Empty() {
		key := rsc.queue.Front()
		rsc.queue.Dequeue()
		go syncReplicaSet(key.(string))
	}

	//go func() {
	//	fmt.Println("[replicaSet Controller] start watch...")
	//	err := rsc.ls.Watch("/testAddPod", rsc.testAddRS, rsc.stopChannel)
	//	if err != nil {
	//		fmt.Printf("[kubelet] watch podConfig error " + err.Error())
	//	} else {
	//		fmt.Println("[kubelet] return...")
	//		return
	//	}
	//	time.Sleep(10 * time.Second)
	//}()
}

func (rsc *ReplicaSetController) syncReplicaSet(key string) error {
	fmt.Println("[ReplicaSet Controller] start sync ...")

	// 获取replicaset对象以及关联的pod对象列表
	if rs, ok := rsc.hashMap.Get(key).(*object.ReplicaSet); !ok {
		fmt.Printf("[ReplicaSet Controller] %v has been deleted/n", key)
		// TODO: 这是什么
		//rsc.expectations.DeleteExpectations(key)
		return nil
	} else {
		// TODO: 判断上一次对replicaset对象的调谐操作中，调用的rsc.manageReplicas方法是否执行完成
		//rsNeedsSync := rsc.expectations.SatisfiedExpectations(key)

		//列出该rs所有的Pods
		var pods []*object.Pod
		podLists, _ := rsc.ls.List(_const.AllPodPrefix)
		for _, eachPod := range podLists {
			pod := &object.Pod{}
			err := json.Unmarshal(eachPod.ValueBytes, &pod)
			if err != nil {
				fmt.Printf("[ReplicaSet Controller] getting pod fail\n")
				break
			}
			// 列出所有有owner且active的pod
			if isOwner(pod.Metadata.OwnerReference, rs.Name, pod.Metadata.Uid) && isActive(pod.Status) {
				pods = append(pods, pod)
			}
		}

		// 调用rsc.manageReplicas增删Pod
		manageReplicasErr := rsc.manageReplicas(pods, rs)

		// 调用calculateStatus计算replicaset的status，并更新
	}
}

func (rsc *ReplicaSetController) testAddRS(res etcdstorage.WatchRes) {

}

func (rsc *ReplicaSetController) manageReplicas(filteredPod []*object.Pod, rs *object.ReplicaSet) error {

}

func isOwner(ownerReferences []object.OwnerReference, name string, UID string) bool {
	for _, owner := range ownerReferences {
		if owner.Name == name && owner.UID == UID {
			return true
		}
	}
	return false
}

func isActive(status object.PodStatus) bool {
	if status.Phase == "Running" {
		return true
	}
	return false
}
