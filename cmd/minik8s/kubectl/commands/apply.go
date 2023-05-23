package commands

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/urfave/cli"
	v2 "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
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
	path := _const.JOBFILE
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

func createNewPod(pod *object.Pod) {
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewRS(rs *object.ReplicaSet) {
	rsRaw, _ := json.Marshal(rs)
	reqBody := bytes.NewBuffer(rsRaw)

	suffix := _const.RS_CONFIG_PREFIX + "/" + rs.Name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewJob(job *object.GPUJob) {
	uuid, err := uuid2.NewUUID()
	if err != nil {
		fmt.Println(err)
	}
	job.Metadata.Uid = uuid.String()
	zip, err := os.ReadFile(job.Spec.App.AppSpec.ZipPath)
	if err != nil {
		fmt.Println(err)
	}
	jobApp := &object.JobAppFile{
		Key:   "job-" + job.Metadata.Uid,
		Slurm: job.NewSlurmScript(),
		App:   zip,
	}

	fmt.Println(jobApp.Slurm)

	//上传job可执行文件到sharedData
	jobAppRaw, _ := json.Marshal(jobApp)
	reqBody := bytes.NewBuffer(jobAppRaw)
	suffix := _const.SHARED_DATA_PREFIX + "/" + jobApp.Key

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)

	//上传job到etcd
	//jobRaw, _ := json.Marshal(job)
	//reqBody2 := bytes.NewBuffer(jobRaw)
	//suffix2 := _const.JOB_CONFIG_PREFIX + "/" + jobApp.Key
	//
	//req2, _ := http.NewRequest("PUT", _const.BASE_URI+suffix2, reqBody2)
	//resp, _ = http.DefaultClient.Do(req2)
	//
	//fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}
