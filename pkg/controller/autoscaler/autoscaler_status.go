package autoscaler

import (
	"Mini-K8s/pkg/controller/autoscaler/resource"
	"Mini-K8s/pkg/object"
)

type resourceStatus struct {
	metadata object.ObjMetadata
	memory   float64
	cpu      float64
}

func (asc *AutoScaleController) getPodResourceStatus(pods []*object.Pod) ([]resourceStatus, error) {
	var statusList []resourceStatus
	for _, pod := range pods {
		var status resourceStatus
		status.metadata = pod.Metadata

		res, err := asc.metricClient.GetResource(resource.CPU, pod.Name, pod.Metadata.Uid)
		if err != nil || res == nil {
			return statusList, err
		}
		status.cpu = *res

		res, err = asc.metricClient.GetResource(resource.MEMORY, pod.Name, pod.Metadata.Uid)
		if err != nil || res == nil {
			return statusList, err
		}
		status.memory = *res

		statusList = append(statusList, status)
	}
	return statusList, nil
}
