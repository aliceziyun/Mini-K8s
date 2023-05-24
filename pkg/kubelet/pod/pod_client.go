package pod

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func updatePod(pod *object.Pod) error {
	pod.Status.Phase = POD_RUNNING_STATUS
	//TODO: get running containers
	pod.Status.RunningContainers = int32(len(pod.Spec.Containers))

	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name
	body, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	reqBody := bytes.NewBuffer(body)
	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}
