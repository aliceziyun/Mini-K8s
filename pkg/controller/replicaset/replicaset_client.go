package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func GetReplicaSetOf(pod *object.Pod, rsc *ReplicaSetController) *object.ReplicaSet {
	ownerReferences := pod.Metadata.OwnerReference
	if len(ownerReferences) == 0 {
		return nil
	}
	for _, owner := range ownerReferences {
		if owner.Kind == object.REPLICASET {
			// 有上层ReplicaSet，查询该pod所属的replicaset是否存在
			suffix := _const.RS_PREFIX + owner.Name
			rsRaw, err := rsc.ls.List(suffix)
			if err != nil {
				fmt.Println("[ReplicaSet Controller] fail to get pod's rs")
				return nil
			}
			rs := &object.ReplicaSet{}
			err = json.Unmarshal(rsRaw[0].ValueBytes, rs)
			if err != nil {
				fmt.Println("[ReplicaSet Controller] get rs with wrong message")
				return nil
			}
			return rs
		}
	}
	return nil
}

func UpdateStatus(rs *object.ReplicaSet) error {
	suffix := _const.RS_PREFIX + rs.Name
	body, err := json.Marshal(rs)
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

func DeleteRS(name string) error {
	suffix := _const.RS_PREFIX + name
	request, err := http.NewRequest("DELETE", _const.BASE_URI+suffix, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("StatusCode not 200")
	}
	return nil
}
