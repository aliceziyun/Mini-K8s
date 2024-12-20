package autoscaler

import (
	"Mini-K8s/pkg/object"
	"fmt"
)

type resourceStatus struct {
	metadata object.ObjMetadata
	name     string
	res      float64
}

// 获取Pod某一项具体的resource的状况
func (asc *AutoScaleController) getPodResourceStatus(resourceName string, pods []*object.Pod) ([]resourceStatus, error) {
	fmt.Println("[AutoScale Controller] get pod resource")
	var statusList []resourceStatus
	for _, pod := range pods {
		//if pod.Spec.NodeName != _const.NODE_NAME {
		//	continue
		//}

		var status resourceStatus
		status.metadata = pod.Metadata
		status.name = resourceName

		switch resourceName {
		case "cpu":
			res, err := asc.metricClient.GetResource("cpu", pod.Name, pod.Metadata.Uid)
			if err != nil || res == nil {
				return statusList, err
			}
			status.res = *res
			break
		case "memory":
			res, err := asc.metricClient.GetResource("memory", pod.Name, pod.Metadata.Uid)
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
