// 暂时用来参考
package kubelet

import (
	// "Mini-K8s/pkg/client"
	// "Mini-K8s/pkg/kubelet/monitor"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/kubelet/podConfig"
	"Mini-K8s/pkg/kubelet/podManager"
	// "Mini-K8s/pkg/kubeproxy"
	// "Mini-K8s/pkg/listerwatcher"
	// "Mini-K8s/pkg/netSupport"
)

type Kubelet struct {
	podManager *podManager.PodManager
	PodConfig  *podConfig.PodConfig
	// podMonitor     *monitor.DockerMonitor
	// kubeNetSupport *netSupport.KubeNetSupport
	// kubeProxy      *kubeproxy.KubeProxy
	// ls          *listerwatcher.ListerWatcher
	// stopChannel <-chan struct{}
	Client client.RESTClient
	Err    error
}

func NewKubelet(lsConfig *listerwatcher.Config, clientConfig string, node *object.Node) *Kubelet {
	kubelet := &Kubelet{}
	kubelet.podManager = podManager.NewPodManager(clientConfig)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig,
	}
	kubelet.Client = restClient

	// initialize list watch
	// ls, err := listerwatcher.NewListerWatcher(lsConfig)
	// if err != nil {
	// 	fmt.Printf("[NewKubelet] list watch start fail...")
	// }
	// kubelet.ls = ls
	// kubelet.kubeNetSupport, err = netSupport.NewKubeNetSupport(lsConfig, clientConfig, node)
	// if err != nil {
	// 	fmt.Printf("[NewKubelet] new kubeNetSupport fail")
	// }
	// kubelet.kubeProxy = kubeproxy.NewKubeProxy(lsConfig, clientConfig)
	// initialize pod podConfig
	kubelet.PodConfig = podConfig.NewPodConfig()

	// kubelet.podMonitor = monitor.NewDockerMonitor()

	return kubelet
}

// func (kl *Kubelet) Run() {
// 	kl.kubeNetSupport.StartKubeNetSupport()
// 	kl.kubeProxy.StartKubeProxy()
// 	updates := kl.PodConfig.GetUpdates()
// 	go kl.podMonitor.Listener()
// 	go kl.syncLoop(updates, kl)
// 	go kl.DoMonitor(context.Background())
// 	go func() {
// 		err := kl.ls.Watch(config.PodConfigPREFIX, kl.watchPod, kl.stopChannel)
// 		if err != nil {
// 			fmt.Printf("[kubelet] watch podConfig error" + err.Error())
// 		} else {
// 			return
// 		}
// 		time.Sleep(10 * time.Second)
// 	}()
// 	go func() {
// 		err := kl.ls.Watch(config.SharedDataPrefix, kl.watchSharedData, kl.stopChannel)
// 		if err == nil {
// 			return
// 		}
// 		time.Sleep(10 * time.Second)
// 	}()
// }

// func (kl *Kubelet) syncLoop(updates <-chan types.PodUpdate, handler SyncHandler) {
// 	for {
// 		kl.syncLoopIteration(updates, handler)
// 	}
// }

// func (kl *Kubelet) AddPod(pod *object.Pod) error {
// 	return kl.podManager.AddPod(pod)
// }
