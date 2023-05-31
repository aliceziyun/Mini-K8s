package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
)

// 对replicaset进行recover，将所有存在的replicaset的name与本体相对应
func (rsc *ReplicaSetController) recover() {
	resList, err := rsc.ls.List(_const.RS_CONFIG_PREFIX)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, res := range resList {
		result := &object.ReplicaSet{}
		err = json.Unmarshal(res.ValueBytes, result)
		rsc.hashMap.Put(result.Name, result)
	}
}
