package replicaset

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
)

// 注意: pod的eventHandler处理逻辑依然是将pod对应的replicaset对象加入queue中，而不是将pod加入到queue中。
func (rsc *ReplicaSetController) handlePod(res etcdstorage.WatchRes) {
	var pod *object.Pod
	var err error
	fmt.Println(res.ResType)
	if res.ResType == etcdstorage.DELETE {
		pod, err = getPodByName(res.Key, rsc.ls)
		fmt.Printf("[ReplicaSet Controller] delete Pod %s \n", pod.Name)
	} else {
		pod = &object.Pod{}
		err = json.Unmarshal(res.ValueBytes, pod)
		fmt.Printf("[ReplicaSet Controller] get new Pod %s \n", pod.Name)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	if pod == nil {
		fmt.Println("[ReplicaSet Controller] pod has been deleted")
		return
	}

	// 获取该pod对应的rs
	rs := getReplicaSetOf(pod, rsc)
	if rs == nil {
		return
	}
	key := getKey(rs)
	rsc.hashMap.Put(key, rs)
	rsc.queue.Enqueue(key)
}

func (rsc *ReplicaSetController) handleRS(res etcdstorage.WatchRes) {
	fmt.Printf("[Replica SetController] get new RS \n")
	if res.ResType == etcdstorage.DELETE {
		return
	}

	rs := &object.ReplicaSet{}
	err := json.Unmarshal(res.ValueBytes, rs)
	if err != nil {
		fmt.Printf("[Replica SetController] get wrong message when handle RS \n")
		return
	}

	// 将key和rs绑定
	key := getKey(rs)
	rsc.hashMap.Put(key, rs)
	rsc.queue.Enqueue(key)
}

func getKey(rs *object.ReplicaSet) string {
	return rs.Name + rs.Uid
}
