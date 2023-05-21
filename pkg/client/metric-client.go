package client

import (
	"Mini-K8s/pkg/controller/autoscaler/resource"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type MetricClient struct {
	Base string
}

// GetResource :获取某个pod的CPU占比情况和内存占比情况
func (mcli *MetricClient) GetResource(resource string, podName string, podUID string) (*float64, error) {
	url, err := mcli.newQuery(resource, podName, podUID)
	if err != nil || url == nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body

	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {

		}
	}(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	metric := QueryRes{}
	err = json.Unmarshal(data, &metric)
	if err != nil {
		return nil, err
	}
	res := metric.Data.ResultArray[0].Value[1]
	v := res.(string)
	resFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, err
	}
	return &resFloat, nil
}

func (mcli *MetricClient) newQuery(rsc string, podName string, podUID string) (*string, error) {
	var resourceTag string
	if rsc == resource.CPU {
		resourceTag = "cpu"
	} else if rsc == resource.MEMORY {
		resourceTag = "memory"
	} else {
		return nil, errors.New("invalid resource")
	}

	query := fmt.Sprintf("query=node_monitor{resource=\"%s\",pod=\"%s\",uid=\"%s\"}", resourceTag, podName, podUID)

	url := mcli.Base + "/api/v1/query?" + query
	return &url, nil
}
