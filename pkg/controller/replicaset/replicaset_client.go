package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

func CreatePod(rs *object.ReplicaSet) error {
	podUID, _ := uuid.NewUUID()
	suffix := _const.ETCD_POD_PREFIX + "/" + rs.Name + podUID.String()

	//get pod infomation
	pod := &object.Pod{}
	pod.Spec = rs.Spec.Pods.Spec
	owner := object.OwnerReference{
		Kind:       object.REPLICASET,
		Name:       rs.Name,
		UID:        rs.Uid,
		Controller: true,
	}
	pod.Metadata.OwnerReference = append(pod.Metadata.OwnerReference, owner)
	podRawData, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRawData)

	// http request
	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}

func DeletePod(podName string) error {
	suffix := _const.POD_CONFIG_PREFIX + podName
	request, err := http.NewRequest("DELETE", _const.BASE_URI+suffix, nil)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}
