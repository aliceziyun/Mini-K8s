package commands

import (
	"Mini-K8s/pkg/object"
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
	pod := object.Pod{}

	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	req, _ := http.NewRequest("PUT", "http://localhost:8080/testpod", reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Println(resp.StatusCode)

	return
}
