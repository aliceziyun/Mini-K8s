package kubelet

import (
	// "Mini-K8s/pkg/client"
	// "Mini-K8s/pkg/kubelet/monitor"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/kubelet/livenessManager"
	"Mini-K8s/pkg/kubelet/podConfig"
	"Mini-K8s/pkg/kubelet/podManager"
	"Mini-K8s/pkg/object"
	"context"
	// "Mini-K8s/pkg/kubeproxy"
	// "Mini-K8s/pkg/listerwatcher"
	// "Mini-K8s/pkg/netSupport"
)

//	type SyncHandler interface {
//		HandlePodAdditions(pods []*v1.Pod)
//		HandlePodUpdates(pods []*v1.Pod)
//		HandlePodRemoves(pods []*v1.Pod)
//		HandlePodReconcile(pods []*v1.Pod)
//		HandlePodSyncs(pods []*v1.Pod)
//		HandlePodCleanups(ctx context.Context) error
//	}
type SyncHandler interface {
	HandlePodAdditions(pods []*object.Pod)
	HandlePodUpdates(pods []*object.Pod)
	HandlePodRemoves(pods []*object.Pod)
	HandlePodSyncs(pods []*object.Pod)           //
	HandlePodCleanups(ctx context.Context) error //
}

type Kubelet struct {
	podManager *podManager.PodManager
	PodConfig  *podConfig.PodConfig
	// podMonitor     *monitor.DockerMonitor
	// kubeNetSupport *netSupport.KubeNetSupport
	// kubeProxy      *kubeproxy.KubeProxy
	// ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	Client      client.RESTClient
	Err         error
	//
	podWorkers      PodWorkers                      //
	livenessManager livenessManager.LivenessManager //

}

type PodWorkers struct {
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

//blog:---------------------------------------
// func (kl *Kubelet) syncLoop(updates <-chan kubetypes.PodUpdate, handler SyncHandler) {
// 	glog.Info("Starting kubelet main sync loop.")
// 	// The resyncTicker wakes up kubelet to checks if there are any pod workers
// 	// that need to be sync'd. A one-second period is sufficient because the
// 	// sync interval is defaulted to 10s.
// 	syncTicker := time.NewTicker(time.Second)
// 	defer syncTicker.Stop()
// 	housekeepingTicker := time.NewTicker(housekeepingPeriod)
// 	defer housekeepingTicker.Stop()
// 	plegCh := kl.pleg.Watch()
// 	const (
// 		base   = 100 * time.Millisecond
// 		max    = 5 * time.Second
// 		factor = 2
// 	)
// 	duration := base
// 	for {
// 		if rs := kl.runtimeState.runtimeErrors(); len(rs) != 0 {
// 			glog.Infof("skipping pod synchronization - %v", rs)
// 			// exponential backoff
// 			time.Sleep(duration)
// 			duration = time.Duration(math.Min(float64(max), factor*float64(duration)))
// 			continue
// 		}
// 		// reset backoff if we have a success
// 		duration = base

// 		kl.syncLoopMonitor.Store(kl.clock.Now())
// 		if !kl.syncLoopIteration(updates, handler, syncTicker.C, housekeepingTicker.C, plegCh) {
// 			break
// 		}
// 		kl.syncLoopMonitor.Store(kl.clock.Now())
// 	}
// }
// syncLoopIteration
