package replicaset

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func createPod(rs *object.ReplicaSet) error {
	podUID, _ := uuid.NewUUID()
	suffix := _const.POD_CONFIG_PREFIX + "/" + rs.Name + podUID.String()

	fmt.Printf("[ReplicaSet Controller] Create New Pod with UID %s... \n", podUID)

	//get pod infomation
	pod := &object.Pod{}
	pod.Name = rs.Name + podUID.String()
	pod.Kind = object.POD
	pod.Spec = rs.Spec.Pods.Spec
	pod.Metadata = rs.Spec.Pods.Metadata
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

func deletePod(name string) error {
	//都已经知道了pod但还是传一个pod的name，这个函数多少有点呆
	nameRaw, _ := json.Marshal(name)
	reqBody := bytes.NewBuffer(nameRaw)
	suffix := _const.POD_CONFIG_PREFIX
	request, err := http.NewRequest("DELETE", _const.BASE_URI+suffix, reqBody)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}

func getReplicaSetOf(pod *object.Pod, rsc *ReplicaSetController) *object.ReplicaSet {
	ownerReferences := pod.Metadata.OwnerReference
	if len(ownerReferences) == 0 {
		return nil
	}
	for _, owner := range ownerReferences {
		if owner.Kind == object.REPLICASET {
			// 有上层ReplicaSet，查询该pod所属的replicaset是否存在
			suffix := _const.RS_CONFIG_PREFIX + "/" + owner.Name
			rsRaw, err := rsc.ls.List(suffix)
			if err != nil {
				fmt.Println("[ReplicaSet Controller]", err)
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

func updateStatus(rs *object.ReplicaSet) error {
	suffix := _const.RS_CONFIG_PREFIX + "/" + rs.Name
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

func deleteRS(name string) error {
	suffix := _const.RS_CONFIG_PREFIX + name
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

func GetAllPods(ls *listwatcher.ListWatcher, name string, UID string) ([]*object.Pod, error) {
	var pods []*object.Pod
	podLists, err := ls.List(_const.POD_CONFIG_PREFIX)
	for _, eachPod := range podLists {
		pod := &object.Pod{}
		err := json.Unmarshal(eachPod.ValueBytes, &pod)
		if err != nil {
			fmt.Printf("[ReplicaSet Controller] getting pod fail \n")
			break
		}
		// 列出所有有owner且active的pod
		if isOwner(pod.Metadata.OwnerReference, name) && isActive(pod.Status) {
			pods = append(pods, pod)
		}
	}
	return pods, err
}

func getPodByName(name string, ls *listwatcher.ListWatcher) (*object.Pod, error) {
	fields := strings.Split(name, "/")
	podName := fields[len(fields)-1]
	suffix := _const.POD_CONFIG_PREFIX + "/" + podName

	podList, err := ls.List(suffix)
	if err != nil {
		fmt.Println(err)
	}
	pod := &object.Pod{}
	if len(podList) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(podList[0].ValueBytes, &pod)
	if err != nil {
		fmt.Println(err)
	}
	return pod, err
}
