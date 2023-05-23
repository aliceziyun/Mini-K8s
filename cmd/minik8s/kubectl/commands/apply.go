package commands

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	v2 "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

func NewApplyCommand() cli.Command {
	applyCmd := cli.Command{
		Name:  "apply",
		Usage: "create pod according to file",
		Action: func(c *cli.Context) error {
			applyFile()
			fmt.Println("apply okk")
			return nil
		},
	}
	return applyCmd
}

func applyFile() {
	path := _const.SERVFILE
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("open file err: %v\n", err)
	}
	mp := make(map[string]any, 2)
	err = v2.Unmarshal(data, mp)

	switch mp["kind"] {
	case "Pod":
		pod := &object.Pod{}
		err = v2.Unmarshal([]byte(data), pod)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewPod(pod)
		break
	case "Service":
		//applyPod1()
		//applyPod2()
		service := &object.Service{}
		err = v2.Unmarshal([]byte(data), service)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewService(service)
		break
	case "ReplicaSet":
		rs := &object.ReplicaSet{}
		err = v2.Unmarshal([]byte(data), rs)
		if err != nil {
			fmt.Println(err)
		}
		createNewRS(rs)
		fmt.Println(rs)
		break
	case "Job":
		job := &object.GPUJob{}
		err = v2.Unmarshal([]byte(data), job)
		if err != nil {
			fmt.Println(err)
		}
		createNewJob(job)
		fmt.Println(job)
		break
	}
	return
}

func applyPod1() {
	path := _const.PODFILE
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("open file err: %v\n", err)
	}
	mp := make(map[string]any, 2)
	err = v2.Unmarshal(data, mp)
	pod := &object.Pod{}
	err = v2.Unmarshal([]byte(data), pod)
	if err != nil {
		fmt.Printf("file in %s unmarshal fail, use default config", path)
	}
	createNewPod(pod)
}

func applyPod2() {
	path := "/home/lcz/go/src/Mini-K8s/build/Pod/testPod2.yaml"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("open file err: %v\n", err)
	}
	mp := make(map[string]any, 2)
	err = v2.Unmarshal(data, mp)
	pod := &object.Pod{}
	err = v2.Unmarshal([]byte(data), pod)
	if err != nil {
		fmt.Printf("file in %s unmarshal fail, use default config", path)
	}
	createNewPod(pod)
}

func createNewPod(pod *object.Pod) {
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name

	req, _ := http.NewRequest("PUT", "http://localhost:8080"+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewService(service *object.Service) {
	fmt.Println(service)
	podRaw, _ := json.Marshal(service)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.SERVICE_CONFIG_PREFIX + "/" + service.Name

	req, _ := http.NewRequest("PUT", "http://localhost:8080"+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewRS(rs *object.ReplicaSet) {
	rsRaw, _ := json.Marshal(rs)
	reqBody := bytes.NewBuffer(rsRaw)

	suffix := _const.RS_CONFIG_PREFIX + "/" + rs.Name

	req, _ := http.NewRequest("PUT", "http://localhost:8080"+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewJob(job *object.GPUJob) {

}
