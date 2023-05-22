package autoscaler

import (
	"Mini-K8s/pkg/controller/autoscaler/resource"
	"Mini-K8s/pkg/object"
)

type resourceStatus struct {
	metadata object.ObjMetadata
	name     string
	res      float64
}

// 获取Pod某一项具体的resource的状况
func (asc *AutoScaleController) getPodResourceStatus(resourceName string, pods []*object.Pod) ([]resourceStatus, error) {
	var statusList []resourceStatus
	for _, pod := range pods {
		var status resourceStatus
		status.metadata = pod.Metadata
		status.name = resourceName

		switch resourceName {
		case "cpu":
			res, err := asc.metricClient.GetResource(resource.CPU, pod.Name, pod.Metadata.Uid)
			if err != nil || res == nil {
				return statusList, err
			}
			status.res = *res
			break
		case "memory":
			res, err := asc.metricClient.GetResource(resource.MEMORY, pod.Name, pod.Metadata.Uid)
			if err != nil || res == nil {
				return statusList, err
			}
			status.res = *res
			break
		}
		statusList = append(statusList, status)
	}
	return statusList, nil
}
