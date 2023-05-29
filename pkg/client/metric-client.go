package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type MetricClient struct {
	Base string
}

// GetResource :获取某个pod的CPU占比情况和内存占比情况
func (mcli *MetricClient) GetResource(resource string, podName string, podUID string) (*float64, error) {
	request, err := http.NewRequest("GET", mcli.Base, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("[Monitor] StatusCode is ", response.StatusCode)
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body

	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {

		}
	}(reader)

	data, err := io.ReadAll(reader)

	resFloat := getData(data, podName, resource)
	if resFloat < 0 {
		err = errors.New("no correspond resource")
	}

	return &resFloat, err
}

func getData(data []byte, rsc string, podName string) float64 {
	p := string(data)
	s := bufio.NewScanner(strings.NewReader(p))
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "pod_metric") {
			res := getResult(s.Text(), podName, rsc)
			if res < 0 {
				continue
			}
			return res
		}
	}
	return -1
}

func getResult(str string, rsc string, name string) float64 {
	//按空格分隔
	arr := strings.Fields(str)
	metricConfig := arr[0]
	metricValue := arr[1]
	start := strings.Index(metricConfig, "{")
	end := strings.Index(metricConfig, "}")
	metricConfig = metricConfig[start+1 : end]

	//分隔成每个字段
	attributes := strings.Split(metricConfig, ",")
	var f = true
	for _, attribute := range attributes {
		if strings.HasPrefix(attribute, "pod") {
			f = f && strings.Contains(attribute, name)
		} else if strings.HasPrefix(attribute, "resource") {
			f = f && strings.Contains(attribute, rsc)
		}
	}
	if f {
		res, _ := strconv.ParseFloat(metricValue, 64)
		return res
	}

	return -1
}
