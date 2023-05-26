package kubelet

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/kubelet/PodUpdate"
	"Mini-K8s/pkg/kubelet/podConfig"
	"Mini-K8s/pkg/kubelet/podManager"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/monitor"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/file"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
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
	podMonitor *monitor.Monitor
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
	kubelet.podMonitor = monitor.NewMonitor()

	return kubelet
}

func (kl *Kubelet) Run() {
	//kl.kubeNetSupport.StartKubeNetSupport()
	//kl.kubeProxy.StartKubeProxy()
	go kl.podMonitor.Listener()
	updates := kl.PodConfig.GetUpdates()
	go kl.syncLoop(updates)
	go kl.monitor(context.Background())

	fmt.Println("[Kubelet] start...")
	stopChan := make(chan int)

	go func() {
		fmt.Println("[Kubelet] start watch pod...")
		err := kl.ls.Watch(_const.POD_CONFIG_PREFIX, kl.watchPod, kl.stopChannel)
		if err != nil {
			fmt.Printf("[Kubelet] watch pod error " + err.Error())
		} else {
			fmt.Println("[Kubelet] return...")
			stopChan <- 1
			return
		}
		time.Sleep(10 * time.Second)
	}()

	go func() {
		fmt.Println("[Kubelet] start watch shared data...")
		err := kl.ls.Watch(_const.SHARED_DATA_PREFIX, kl.watchSharedData, kl.stopChannel)
		if err != nil {
			fmt.Printf("[Kubelet] watch shared_data error " + err.Error())
		} else {
			fmt.Println("[Kubelet] return...")
			stopChan <- 1
			return
		}
		time.Sleep(10 * time.Second)
	}()

	<-stopChan
}

func (kl *Kubelet) syncLoop(ch <-chan PodUpdate.PodUpdate) bool {
	fmt.Println("[Kubelet] start syncLoop...")
	for {
		select {
		case u, open := <-ch:
			if !open {
				fmt.Printf("Update channel is closed")
				return false
			}
			fmt.Println("[Kubelet] new coming pod message...")
			switch u.Op {
			case ADD:
				kl.HandlePodAdd(u.Pods)
				break
			case DELETE:
				kl.HandlePodDelete(u.Pods)
				break
			case UPDATE:
				kl.HandlePodUpdates(u.Pods)
				break
			}

		}
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

func (kl *Kubelet) HandlePodDelete(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] delete pod:%s \n", pod.Name)
		err := kl.podManager.DeletePod(pod.Name)
		if err != nil {
			fmt.Printf("[Kubelet] Delete pod fail...\n")
		}
	}
}

func (kl *Kubelet) HandlePodUpdates(pods []*object.Pod) {
	for _, pod := range pods { //先删除
		err := kl.podManager.DeletePod(pod.Name)
		if err != nil {
			fmt.Printf("[Kubelet] Delete pod fail...")
			fmt.Printf(err.Error())
			kl.Err = err
		}
	}
	for _, pod := range pods { //再创建
		err := kl.podManager.AddPod(pod)
		if err != nil {
			fmt.Printf("[Kubelet] Add pod fail...")
			fmt.Printf(err.Error())
			kl.Err = err
		}
		fmt.Printf("[Kubelet] update pod %s \n", pod.Name)
	}
}

func (kl *Kubelet) watchPod(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Println("[kubelet]", err)
	}

	fmt.Println("[kubelet] Add Pod")
	pods := []*object.Pod{pod}
	//检查pod是否已经存在
	ok := kl.podManager.CheckIfPodExist(pod.Name)
	if !ok { //pod不存在
		if pod.Status.Phase != object.DELETED {
			fmt.Printf("[Kubelet] create new pod %s ! \n", pod.Name)
			//新建
			podUp := PodUpdate.PodUpdate{
				Pods: pods,
				Op:   ADD,
			}
			kl.PodConfig.GetUpdates() <- podUp
		}
	} else { //pod已经存在
		fmt.Printf("[Kubelet] pod %s exists ! \n", pod.Name)
		if pod.Status.Phase == object.DELETED {
			//删除pod
			podUp := PodUpdate.PodUpdate{
				Pods: pods,
				Op:   DELETE,
			}
			kl.PodConfig.GetUpdates() <- podUp
		} else {
			//更新pod
			podUp := PodUpdate.PodUpdate{
				Pods: pods,
				Op:   UPDATE,
			}
			kl.PodConfig.GetUpdates() <- podUp
		}
	}

}

// 每隔10秒更新一次pod的状态
func (kl *Kubelet) monitor(ctx context.Context) {
	for {
		fmt.Printf("[Kubelet] New round monitoring...\n")
		podMap := kl.podManager.CopyName2pod()
		for _, pod := range podMap {
			kl.podMonitor.GetDockerStat(ctx, pod)
		}
		time.Sleep(time.Second * 10)
	}
}

// 查看sharedData
func (kl *Kubelet) watchSharedData(res etcdstorage.WatchRes) {
	switch res.ResType {
	case etcdstorage.PUT:
		fmt.Println("[Kubelet] new shared data...")
		jobAppFile := object.JobAppFile{}
		err := json.Unmarshal(res.ValueBytes, &jobAppFile)
		if err != nil {
			fmt.Println("[Kubelet]", err)
			return
		}
		appName := jobAppFile.Key + ".zip"
		unzippedDir := path.Join(_const.SHARED_DATA_DIR, jobAppFile.Key)

		//将文件放入对应位置
		err = file.Bytes2File(jobAppFile.App, appName, _const.SHARED_DATA_DIR)
		if err != nil {
			fmt.Println("[Kubelet]", err)
			return
		}
		err = file.Unzip(path.Join(_const.SHARED_DATA_DIR, appName), unzippedDir)
		if err != nil {
			fmt.Println("[Kubelet]", err)
			return
		}
		err = file.Bytes2File(jobAppFile.Slurm, "sbatch.slurm", unzippedDir)
		if err != nil {
			fmt.Println("[Kubelet]", err)
			return
		}

		fmt.Println("[kubelet] Add Shared Data")
		break
	}
}
