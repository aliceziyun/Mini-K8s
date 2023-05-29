package commands

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/printer"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"strings"
	"time"
)

var resourceList = []string{object.POD, object.REPLICASET, object.SERVICE, object.HPA}

func NewGetPodCommand() cli.Command {
	getPodCmd := cli.Command{
		Name:  "get",
		Usage: "view lists of resources such as pod",
		Action: func(c *cli.Context) error {
			if len(c.Args()) > 2 {
				fmt.Println("Only one resource object can be requested at a time!")
				return nil
			}
			if len(c.Args()) <= 2 {
				newGetAllRequest(c.Args().Get(0))
			}
			return nil
		},
	}
	return getPodCmd
}

func newGetAllRequest(arg string) {
	if arg == "" {
		fmt.Println("You need to specify a resource object!")
		fmt.Printf("[Possible Resource Object]: ")
		printPossibleResourceObj()
		return
	}
	switch arg {
	case strings.ToLower(object.POD):
		printer.PrintPods(getPods())
		return
	case strings.ToLower(object.REPLICASET):
		printer.PrintRS(getRS())
	case strings.ToLower(object.NODE):
		printer.PrintNode(getNodes())
	default:
		fmt.Println("No such resource!")
		fmt.Printf("[Possible Resource Object]: ")
		printPossibleResourceObj()
		return
	}
}

func getPods() []object.UserPod {
	resList := getAll(_const.BASE_URI + _const.POD_RUNTIME_PREFIX)
	var usrPods []object.UserPod
	for _, res := range resList {
		pod := &object.Pod{}
		err := json.Unmarshal(res.ValueBytes, pod)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		ready := fmt.Sprintf("%d/%d", len(pod.Spec.Containers), pod.Status.RunningContainers)
		ctime := time.Now().Sub(pod.Metadata.Ctime).String()
		usrPod := object.UserPod{
			Name:   pod.Name,
			Ready:  ready,
			Status: pod.Status.Phase,
			Owner:  pod.Spec.NodeName,
			IP:     pod.Status.PodIP,
			Ctime:  ctime,
		}
		usrPods = append(usrPods, usrPod)
	}
	return usrPods
}

func getRS() []object.UserRS {
	resList := getAll(_const.BASE_URI + _const.RS_CONFIG_PREFIX)
	var usrRSs []object.UserRS
	for _, res := range resList {
		rs := &object.ReplicaSet{}
		err := json.Unmarshal(res.ValueBytes, rs)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if rs.Status.Status == object.DELETED {
			continue
		}

		usrRS := object.UserRS{
			Name:    rs.Name,
			Ready:   rs.Spec.Replicas,
			Current: rs.Status.ReplicaStatus,
		}
		usrRSs = append(usrRSs, usrRS)
	}
	return usrRSs
}

func getNodes() []object.UserNode {
	resList := getAll(_const.BASE_URI + _const.NODE_CONFIG_PREFIX)
	var usrNodes []object.UserNode
	for _, res := range resList {
		node := &object.Node{}
		err := json.Unmarshal(res.ValueBytes, node)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		ctime := time.Now().Sub(node.MetaData.Ctime).String()
		//ctime = timer.FormatTime(ctime)

		usrNode := object.UserNode{
			Name:      node.MetaData.Name,
			DynamicIP: node.Spec.DynamicIp,
			Role:      node.Roles,
			Ctime:     ctime,
		}
		usrNodes = append(usrNodes, usrNode)
	}
	return usrNodes
}

func printPossibleResourceObj() {
	for _, obj := range resourceList {
		fmt.Printf(strings.ToLower(obj) + " ")
	}
	fmt.Printf("\n")
}

func getAll(url string) []etcdstorage.ListRes {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println(errors.New("StatusCode not 200"))
		return nil
	}
	reader := response.Body
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(reader)
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var resList []etcdstorage.ListRes
	err = json.Unmarshal(data, &resList)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return resList
}
