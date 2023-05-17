package commands

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/third_party/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"net/http"
)

func NewApplyCommand() cli.Command {
	applyCmd := cli.Command{
		Name:  "apply",
		Usage: "create pod according to file",
		Action: func(c *cli.Context) error {
			CreatePod()
			fmt.Println("apply okk")
			return nil
		},
	}
	return applyCmd
}

func CreatePod() {
	pod := util.GetPodByFile(_const.PODFILE)

	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	req, _ := http.NewRequest("PUT", "http://localhost:8080/testAddPod", reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Println(resp.StatusCode)

	return
}
