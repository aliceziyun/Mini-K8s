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
	path := _const.RSFILE
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
	case "ReplicaSet":
		rs := &object.ReplicaSet{}
		err = v2.Unmarshal([]byte(data), rs)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewRS(rs)
		fmt.Println(rs)
		break
	}

	return
}

func createNewPod(pod *object.Pod) {
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name

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
