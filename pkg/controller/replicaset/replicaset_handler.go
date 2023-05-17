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

	// 判断该pod是否有上层controller,有则调用查询该pod所属的replicaset是否存在

}

func (rsc *ReplicaSetController) handleRS(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}

	rs := &object.ReplicaSet{}
	err := json.Unmarshal(res.ValueBytes, rs)
	if err != nil {
		fmt.Printf("[ReplicaSetController] get wrong message when handle RS \n")
		return
	}

	fmt.Printf("[ReplicaSetController] get new RS \n")

	// 将key和rs绑定
	key := getKey(rs)
	rsc.hashMap.Put(key, rs)
	rsc.queue.Enqueue(key)
}

func getKey(rs *object.ReplicaSet) string {
	return rs.Name + rs.Uid
}
