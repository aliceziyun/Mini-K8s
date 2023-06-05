package scheduler

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

var globalCount = 0 //保证第一个node分配到master上，方便测试

// 使用Round Robin
func selectNode(nodes []object.Node) (string, error) {
	num := len(nodes)
	if num == 0 {
		return "", errors.New("[Scheduler] no node to select")
	}
	idx := globalCount % num
	globalCount++

	return nodes[idx].MetaData.Name, nil
}

func selectNodeWithRand(nodes []object.Node) (string, error) {
	num := len(nodes)
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(num)
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
