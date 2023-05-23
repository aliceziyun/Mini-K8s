package selector

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"time"
)

type ServRuntime struct {
	service *object.Service
	ls      listwatcher.ListWatcher
	pods    []*object.Pod
}

func (r *ServRuntime) podSelector() {
	selector := r.service.Spec.Selector

	res, err := r.ls.List("PodRuntimePrefix")
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
		if pod.Status.Phase != "RUNNING" {
			continue
		}

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

	r.pods = selectedPods

	if len(r.pods) == 0 {
		// ERROR! no pod has been selected
		return
	}

	var newPodsInfo []object.PodNameAndIp
	for _, val := range r.pods {
		newPodsInfo = append(newPodsInfo, object.PodNameAndIp{Name: val.Name, Ip: val.Status.PodIP})
	}
	r.service.Spec.PodNameAndIps = newPodsInfo

	data, _ := json.Marshal(r.service)

	store, _ := etcdstorage.InitKVStore([]string{"localhost:2379"}, 5*time.Second)
	store.Put("/service", string(data))
}

func NewService(serv *object.Service, ls listwatcher.ListWatcher) *ServRuntime {
	servRuntime := &ServRuntime{}
	servRuntime.service = serv
	servRuntime.ls = ls
	servRuntime.podSelector()
	// TODO need to watch the change of pods
	return servRuntime
}
