package selector

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"encoding/json"
)

type ServRuntime struct {
	service *object.Service
	ls      *listwatcher.ListWatcher
	Pods    []*object.Pod
}

func (r *ServRuntime) podSelector() {
	selector := r.service.Spec.Selector

	res, err := r.ls.List(_const.POD_RUNTIME_PREFIX)
	if err != nil {
		return
	}

	var allPods []*object.Pod
	for _, val := range res {
		pod := &object.Pod{}
		json.Unmarshal(val.ValueBytes, pod)
		allPods = append(allPods, pod)
	}

	var selectedPods []*object.Pod
	for _, pod := range allPods {

		isSelected := true
		for key, val := range selector {
			podVal, ok := pod.Metadata.Labels[key]
			if !ok || val != podVal {
				isSelected = false
				break
			}
		}

		if isSelected {
			selectedPods = append(selectedPods, pod)
		}
	}

	r.Pods = selectedPods

	if len(r.Pods) == 0 {
		// ERROR! no pod has been selected
		return
	}

	var newPodsInfo []object.PodNameAndIp
	for _, val := range r.Pods {
		newPodsInfo = append(newPodsInfo, object.PodNameAndIp{Name: val.Name, Ip: val.Status.PodIP})
	}
	r.service.Spec.PodNameAndIps = newPodsInfo

}

func NewService(serv *object.Service, lsc *listwatcher.Config) *ServRuntime {
	servRuntime := &ServRuntime{}
	servRuntime.service = serv
	servRuntime.ls, _ = listwatcher.NewListWatcher(lsc)
	servRuntime.podSelector()
	// TODO need to watch the change of Pods
	return servRuntime
}
