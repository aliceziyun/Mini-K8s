package commands

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/urfave/cli"
	v2 "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	path2 "path"
)

var fileName string

func NewApplyCommand() cli.Command {
	applyCmd := cli.Command{
		Name:    "apply",
		Usage:   "create pod according to file",
		Aliases: []string{"a"},
		Action: func(c *cli.Context) error {
			err := applyFile(fileName)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("apply okk")
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "f",
				Value:       "",
				Usage:       "read config from file",
				Required:    true,
				Destination: &fileName,
			},
		}}
	return applyCmd
}

func applyFile(file string) error {
	if file == "" {
		return errors.New("you must include file name")
	}
	filename := file + ".yaml"
	path := path2.Join(_const.WORK_DIR, filename)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("open file err: %v\n", err)
	}
	mp := make(map[string]any, 2)
	err = v2.Unmarshal(data, mp)
	if err != nil {
		return err
	}

	switch mp["kind"] {
	case "Pod":
		pod := &object.Pod{}
		err = v2.Unmarshal([]byte(data), pod)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
			return err
		}
		createNewPod(pod)
		break
	case "Service":
		service := &object.Service{}
		err = v2.Unmarshal([]byte(data), service)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewService(service)
		break
	case "DNS":
		dnsConfig := &object.DNSConfig{}
		err = v2.Unmarshal([]byte(data), dnsConfig)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewDNS(dnsConfig)
		break
	case "ReplicaSet":
		rs := &object.ReplicaSet{}
		err = v2.Unmarshal([]byte(data), rs)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewRS(rs)
		break
	case "HorizontalPodAutoScale":
		hpa := &object.Autoscaler{}
		err = v2.Unmarshal([]byte(data), hpa)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewHPA(hpa)
		break
	case "Job":
		job := &object.GPUJob{}
		err = v2.Unmarshal([]byte(data), job)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewJob(job)
		break
	case "Node":
		node := &object.Node{}
		err = v2.Unmarshal([]byte(data), node)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewNode(node)
		break
	case "Function":
		function := &object.Function{}
		err = v2.Unmarshal([]byte(data), function)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewFunc(function)
		break
	case "Workflow":
		workflow := &object.WorkFlow{}
		err = v2.Unmarshal([]byte(data), workflow)
		if err != nil {
			fmt.Println(err)
			return err
		}
		createNewWorkflow(workflow)
	case "VService":
		vs := &object.VService{}
		err = v2.Unmarshal([]byte(data), vs)
		if err != nil {
			fmt.Printf("file in %s unmarshal fail, use default config", path)
		}
		createNewVService(vs)
		break
	default:
		err = errors.New("no such resources")
		return err
	}
	return nil
}

func createNewPod(pod *object.Pod) {
	fmt.Println("[Kubectl] create new pod ...")
	name := pod.Name
	pod.Name = name

	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.POD_CONFIG_PREFIX + "/" + name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewService(service *object.Service) {
	fmt.Println(service)
	podRaw, _ := json.Marshal(service)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.SERVICE_CONFIG_PREFIX + "/" + service.Name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewDNS(dnsConfig *object.DNSConfig) {
	podRaw, _ := json.Marshal(dnsConfig)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.DNS_CONFIG_PREFIX + "/" + dnsConfig.Name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewVService(vs *object.VService) {
	podRaw, _ := json.Marshal(vs)
	reqBody := bytes.NewBuffer(podRaw)

	suffix := _const.VSERVICE_CONFIG_PREFIX + "/" + vs.Name

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

func createNewHPA(hpa *object.Autoscaler) {
	hpaRaw, _ := json.Marshal(hpa)
	reqBody := bytes.NewBuffer(hpaRaw)

	suffix := _const.HPA_CONFIG_PREFIX + "/" + hpa.Metadata.Name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewNode(node *object.Node) {
	nodeRaw, _ := json.Marshal(node)
	reqBody := bytes.NewBuffer(nodeRaw)

	suffix := _const.NODE_CONFIG_PREFIX + "/" + node.MetaData.Name

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

	//上传job可执行文件到sharedData
	jobAppRaw, _ := json.Marshal(jobApp)
	reqBody := bytes.NewBuffer(jobAppRaw)
	suffix := _const.SHARED_DATA_PREFIX + "/" + jobApp.Key

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)

	//上传job到etcd
	job.Status = object.PD
	jobRaw, _ := json.Marshal(job)
	reqBody2 := bytes.NewBuffer(jobRaw)
	suffix2 := _const.JOB_CONFIG_PREFIX + "/" + jobApp.Key

	req2, _ := http.NewRequest("PUT", _const.BASE_URI+suffix2, reqBody2)
	resp, _ = http.DefaultClient.Do(req2)

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewFunc(funtion *object.Function) {
	funcRaw, err := json.Marshal(funtion)
	if err != nil {
		fmt.Println(err)
		return
	}
	reqBody := bytes.NewBuffer(funcRaw)
	suffix := _const.FUNC_CONFIG_PREFIX + "/" + funtion.Name

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}

func createNewWorkflow(workflow *object.WorkFlow) {
	workflowRaw, err := json.Marshal(workflow)
	if err != nil {
		fmt.Println(err)
		return
	}
	reqBody := bytes.NewBuffer(workflowRaw)
	suffix := _const.WORKFLOW_CONFIG_PREFIX + "/" + workflow.Name

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[kubectl] send request to server with code %d", resp.StatusCode)
}
