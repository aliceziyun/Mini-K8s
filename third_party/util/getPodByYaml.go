package util

import (
	"Mini-K8s/pkg/object"
	"fmt"
	"io/ioutil"

	v2 "gopkg.in/yaml.v2"
)

func GetPodByFile(path string) *object.Pod {
	return getPodByFile(path)
}

func getPodByFile(path string) *object.Pod {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("open file err: %v\n", err)
		return nil
	}
	pod := &object.Pod{}
	err = v2.Unmarshal([]byte(data), pod)
	if err != nil {
		fmt.Printf("file in %s unmarshal fail, use default config", path)
		return nil
	}
	return pod
}
