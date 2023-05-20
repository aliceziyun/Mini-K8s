package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/controller/replicaset/RSConfig"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const (
	BurstReplicas = 10
)

type ReplicaSetController struct {
	ls          *listwatcher.ListWatcher
	config      *RSConfig.RSConfig
	stopChannel <-chan struct{}
	queue       queue.ConcurrentQueue
	hashMap     *_map.ConcurrentMap
}

func NewReplicaSetController(controllerContext controller_context.ControllerContext) *ReplicaSetController {
	rsConfig := RSConfig.NewRSConfig()
	hash := _map.NewConcurrentMap()
	rsc := &ReplicaSetController{
		ls:      controllerContext.Ls,
		config:  rsConfig,
		hashMap: hash,
	}
	return rsc
}

func (rsc *ReplicaSetController) Run(ctx context.Context) {
	fmt.Println("[ReplicaSet Controller] start run ...")

	go rsc.register()
	go rsc.worker() //单worker，足够

}

// 绑定watch，当资源发生变化时，通知controller
func (rsc *ReplicaSetController) register() {
	// register RS handler
	go func() {
		err := rsc.ls.Watch(_const.RS_CONFIG_PREFIX, rsc.handleRS, rsc.stopChannel)
		if err != nil {
			fmt.Println("[ReplicaSet Controller] list watch RS handler init fail...")
		}
	}()

	//register Pod handler
	go func() {
		err := rsc.ls.Watch(_const.POD_CONFIG_PREFIX, rsc.handlePod, rsc.stopChannel)
		if err != nil {
			fmt.Println("[ReplicaSet Controller] list watch Pod handler init fail...")
		}
	}()
}

func (rsc *ReplicaSetController) worker() {
	var m sync.Mutex //这个锁很重要，保证一次同步操作是原子的
	for {
		// 接受到新消息，开始处理
		if !rsc.queue.Empty() {
			key := rsc.queue.Front()
			rsc.queue.Dequeue()
			m.Lock()
			err := rsc.syncReplicaSet(key.(string))
			if err != nil {
				fmt.Println("[ReplicaSet Controller] worker error")
			}
			m.Unlock()
		} else {
			time.Sleep(time.Second)
		}
	}

}

func (rsc *ReplicaSetController) syncReplicaSet(key string) error {
	// TODO: 如果同步Pod状态还比较好理解，如果删除或者增加RS要怎么操作呢
	fmt.Println("[ReplicaSet Controller] start sync ...")

	// 获取replicaset对象以及关联的pod对象列表
	if rs, ok := rsc.hashMap.Get(key).(*object.ReplicaSet); !ok {
		fmt.Printf("[ReplicaSet Controller] %v has been deleted \n", key)
		// TODO: 这是什么
		//rsc.expectations.DeleteExpectations(key)
		return nil
	} else {
		// TODO: 判断上一次对replicaset对象的调谐操作中，调用的rsc.manageReplicas方法是否执行完成
		//rsNeedsSync := rsc.expectations.SatisfiedExpectations(key)

		//列出该rs所有的Pods
		var pods []*object.Pod
		podLists, _ := rsc.ls.List(_const.POD_CONFIG_PREFIX)
		fmt.Println(len(podLists))
		for _, eachPod := range podLists {
			pod := &object.Pod{}
			err := json.Unmarshal(eachPod.ValueBytes, &pod)
			if err != nil {
				fmt.Printf("[ReplicaSet Controller] getting pod fail \n")
				break
			}
			// 列出所有有owner且active的pod
			if isOwner(pod.Metadata.OwnerReference, rs.Name) && isActive(pod.Status) {
				pods = append(pods, pod)
			}
		}

		fmt.Println(len(pods))

		//调用rsc.manageReplicas增删Pod
		err := rsc.manageReplicas(pods, rs)
		if err != nil {
			fmt.Println("[ReplicaSet Controller] manageReplicas fail!")
		}

		// 调用calculateStatus计算replicaset的status，并更新
		newStatus := rsc.calculateStatus(rs, pods)
		_, statusErr := rsc.updateReplicaSetStatus(rs, newStatus)
		return statusErr
	}
	return nil
}

func (rsc *ReplicaSetController) manageReplicas(filteredPods []*object.Pod, rs *object.ReplicaSet) error {
	diff := len(filteredPods) - int(rs.Spec.Replicas)
	if diff == 0 {
		return nil
	}
	if diff < 0 {
		diff *= -1
		// 超过了一次最多可以创建的数量上限，修正
		if diff > BurstReplicas {
			diff = BurstReplicas
		}

		// TODO: 将本轮调谐期望的Pod数量设置进expectation
		//rsc.expectations.ExpectCreations(rsKey, diff)

		//原来K8s是指数级增长创建Pod，现在直接循环创建
		successfulCreations, err := rsc.slowStartBatch(diff, rs)
		//有些pod没有创建成功，下次再创建
		if skippedPods := diff - successfulCreations; skippedPods > 0 {
			fmt.Printf("[ReplicaSet Controller] creation skipped of %d pods \n", skippedPods)
			for i := 0; i < skippedPods; i++ {
				//TODO:补上expectation
				//rsc.expectations.CreationObserved(rsKey)
			}
		}
		return err
	} else {
		if diff > BurstReplicas {
			diff = BurstReplicas
		}
		podsToDelete := filteredPods[:diff]
		//TODO:expectation
		//rsc.expectations.ExpectDeletions(rsKey, getPodKeys(podsToDelete))
		for _, pod := range podsToDelete {
			if err := DeletePod(pod.Name); err != nil {
				//podKey := controller.PodKey(targetPod)
				//rsc.expectations.DeletionObserved(rsKey, podKey)
				fmt.Printf("[ReplicaSet Controller] deletion skipped of pods %v \n", pod.Name)
			}
		}
	}
	return nil
}

// 计算并返回replicaset对象的status
func (rsc *ReplicaSetController) calculateStatus(rs *object.ReplicaSet, filteredPods []*object.Pod) object.ReplicaSetStatus {
	newStatus := rs.Status
	newStatus.ReplicaStatus = int32(len(filteredPods))
	return newStatus
}

// 判断新计算出来的status是否与现存replicaset对象的status中的一致
func (rsc *ReplicaSetController) updateReplicaSetStatus(rs *object.ReplicaSet, newStatus object.ReplicaSetStatus) (*object.ReplicaSet, error) {
	if rs.Status.ReplicaStatus == newStatus.ReplicaStatus {
		return rs, nil
	}
	rs.Status = newStatus
	var err error
	if rs.Status.ReplicaStatus == 0 {
		err = DeleteRS(rs.Name)
	} else {
		err = UpdateStatus(rs)
	}
	return rs, err
}

func (rsc *ReplicaSetController) slowStartBatch(diff int, rs *object.ReplicaSet) (int, error) {
	var success int
	for i := 0; i < diff; i++ {
		err := CreatePod(rs)
		if err == nil {
			success++
		}
	}
	return success, nil
}

func isOwner(ownerReferences []object.OwnerReference, name string) bool {
	//TODO: 疑似需要对UID进行判断
	for _, owner := range ownerReferences {
		if owner.Name == name {
			return true
		}
	}
	return false
}

func isActive(status object.PodStatus) bool {
	//if status.Phase == "Running" {
	return true
	//}
	//return false
}
