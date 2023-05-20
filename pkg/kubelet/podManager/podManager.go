// 暂时用来参考
package podManager

import (
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/kubelet/pod"
	"Mini-K8s/pkg/object"
	"errors"
	"fmt"
)

// 定时更新间隔
const PODMANAGER_TIME_INTERVAL = 20

// 存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	podByName map[string]*pod.Pod //name-pod的映射
	// //对map的保护
	// lock         sync.Mutex
	// client       client.RESTClient
	// clientConfig client.Config
	// Err          error
}

// var instance *PodManager

func NewPodManager(clientConfig client.Config) *PodManager {
	newManager := &PodManager{}
	newManager.podByName = make(map[string]*pod.Pod)
	// restClient := client.RESTClient{
	// 	Base: "http://" + clientConfig.Host,
	// }
	// newManager.client = restClient
	// newManager.clientConfig = clientConfig
	// var lock sync.Mutex
	// newManager.lock = lock
	return newManager
}
func (p *PodManager) CheckPodExist(podName string) bool {
	_, keyExist := p.podByName[podName]
	return keyExist
}

func (p *PodManager) AddPod(config *object.Pod) error {
	// p.lock.Lock()
	// defer p.lock.Unlock()
	//确保pod还不存在，才能创建新的
	if p.CheckPodExist(config.Metadata.Name) {
		return errors.New(config.Metadata.Name + "pod已存在，需先删除原有pod")
	}
	// newPod := pod.NewPodfromConfig(config, p.clientConfig)
	newPod := pod.NewPodfromConfig(config)
	p.podByName[config.Name] = newPod
	return nil
}

func (p *PodManager) DeletePod(podName string) error {
	// p.lock.Lock()
	// defer p.lock.Unlock()
	if !p.CheckPodExist(podName) {
		//不存在该pod
		return errors.New(podName + "对应的pod不存在")
	}
	pod, _ := p.podByName[podName]
	fmt.Printf("[DeleteRuntimePod] Prepare delete pod")
	pod.DeletePod()
	delete(p.podByName, podName) //删除podName为key的pod
	return nil
}

// // CopyName2pod only copy the pointers in map, check before actual use
// func (p *PodManager) CopyName2pod() map[string]*pod.Pod {
// 	p.lock.Lock()
// 	defer p.lock.Unlock()
// 	uuidMap := make(map[string]*pod.Pod)
// 	for key, val := range p.podByName {
// 		uuidMap[key] = val
// 	}
// 	return uuidMap
// }
