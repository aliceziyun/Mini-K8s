package kubelet

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/kubelet/PodUpdate"
	"Mini-K8s/pkg/kubelet/podConfig"
	"Mini-K8s/pkg/kubelet/podManager"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	// SET is the current pod configuration.
	SET = iota
	// ADD signifies pods that are new to this source.
	ADD
	// DELETE signifies pods that are gracefully deleted from this source.
	DELETE
	// UPDATE signifies pods have been updated in this source.
	UPDATE
)

type Kubelet struct {
	podManager *podManager.PodManager
	PodConfig  *podConfig.PodConfig
	// podMonitor     *monitor.DockerMonitor
	// kubeNetSupport *netSupport.KubeNetSupport
	// kubeProxy      *kubeproxy.KubeProxy
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
	Client      client.RESTClient
	Err         error
}

func NewKubelet(lsConfig *listwatcher.Config, clientConfig client.Config) *Kubelet {
	kubelet := &Kubelet{}
	// initialize rest client
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	kubelet.Client = restClient

	// initialize pod manager
	kubelet.podManager = podManager.NewPodManager(clientConfig)

	// initialize list watch
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Println("[NewKubelet] list watch start fail...")
		os.Exit(0)
	}
	kubelet.ls = ls
	kubelet.PodConfig = podConfig.NewPodConfig()

	return kubelet
}

func (kl *Kubelet) Run() {
	//kl.kubeNetSupport.StartKubeNetSupport()
	//kl.kubeProxy.StartKubeProxy()
	//go kl.podMonitor.Listener()
	updates := kl.PodConfig.GetUpdates()
	go kl.syncLoop(updates)
	//go kl.DoMonitor(context.Background())

	fmt.Println("[kubelet] start...")
	ch := make(chan int)

	go func() {
		fmt.Println("[kubelet] start watch...")
		err := kl.ls.Watch(_const.POD_CONFIG_PREFIX, kl.AddPod, kl.stopChannel)
		if err != nil {
			fmt.Printf("[kubelet] watch podConfig error " + err.Error())
		} else {
			fmt.Println("[kubelet] return...")
			ch <- 1
			return
		}
		time.Sleep(10 * time.Second)
	}()

	<-ch
}

func (kl *Kubelet) syncLoop(ch <-chan PodUpdate.PodUpdate) bool {
	fmt.Println("[kubelet] start syncLoop...")
	for {
		select {
		case u, open := <-ch:
			if !open {
				fmt.Printf("Update channel is closed")
				return false
			}
			switch u.Op {
			case ADD:
				kl.HandlePodAdd(u.Pods)
				break
			}
		}
		return true
	}
}

func (kl *Kubelet) HandlePodAdd(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] Prepare add pod:%s\npod:%+v\n", pod.Name, pod)
		err := kl.podManager.AddPod(pod)
		if err != nil {
			fmt.Println("[kubelet]AddPod error" + err.Error())
			kl.Err = err
		}
	}
}

func (kl *Kubelet) AddPod(res etcdstorage.WatchRes) {
	fmt.Println("test Add Pod success")
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Println("[kubelet] watch /testAddPod error", err)
	}
	fmt.Println("[kubelet] /testAddPod new message")
	pods := []*object.Pod{pod}
	podUp := PodUpdate.PodUpdate{
		Pods: pods,
		Op:   ADD,
	}
	kl.PodConfig.GetUpdates() <- podUp
}
