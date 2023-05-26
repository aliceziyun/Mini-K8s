package scheduler

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

var globalCount int
var mtx sync.Mutex

// 使用Round Robin
func selectNode(nodes []object.Node) (string, error) {
	mtx.Lock()
	num := len(nodes)
	if num == 0 {
		return "", errors.New("[Scheduler] no node to select")
	}
	idx := globalCount % num
	globalCount++

	defer mtx.Unlock()

	return nodes[idx].MetaData.Name, nil
}

func updatePod(pod *object.Pod) error {
	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name

	podRawData, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRawData)

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}
