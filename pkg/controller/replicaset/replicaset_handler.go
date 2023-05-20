package replicaset

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
)

// 注意: pod的eventHandler处理逻辑依然是将pod对应的replicaset对象加入queue中，而不是将pod加入到queue中。
func (rsc *ReplicaSetController) handlePod(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}

	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("[ReplicaSet Controller] get wrong message when handle Pod \n")
		return
	}
	fmt.Printf("[ReplicaSet Controller] get new Pod \n")

	// 获取该pod对应的rs
	rs := GetReplicaSetOf(pod, rsc)
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
