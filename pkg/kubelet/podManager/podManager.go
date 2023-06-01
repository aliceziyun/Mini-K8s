package podManager

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/kubelet/pod"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// 定时更新间隔
const PODMANAGER_TIME_INTERVAL = 20

// 存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	name2pod map[string]*pod.Pod //name-pod的映射
	//对map的保护
	lock         sync.Mutex
	client       client.RESTClient
	clientConfig client.Config
	Err          error
}

func NewPodManager(clientConfig client.Config) *PodManager {
	newManager := &PodManager{}
	newManager.name2pod = make(map[string]*pod.Pod)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newManager.client = restClient
	newManager.clientConfig = clientConfig
	var lock sync.Mutex
	newManager.lock = lock

	newManager.recover()

	return newManager
}

// 进行恢复，填充name2pod
func (p *PodManager) recover() {
	resList, err := list(_const.POD_META_PREFIX)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, res := range resList {
		result := &object.PodMeta{}
		err = json.Unmarshal(res.ValueBytes, result)
		if err != nil {
			fmt.Println(err)
			return
		}
		if result.NodeName == _const.NODE_NAME {
			fmt.Println("[Pod Manager] recover with pod", result.PodName)
			res, err := list(_const.POD_RUNTIME_PREFIX + "/" + result.PodName)
			if err != nil {
				fmt.Println(err)
				return
			}
			configPod := &object.Pod{}
			if len(res) == 0 {
				return
			}
			err = json.Unmarshal(res[0].ValueBytes, configPod)
			if err != nil {
				fmt.Println(err)
				return
			}
			pod := pod.RecoverPod(result, configPod)
			p.name2pod[result.PodName] = pod
			pod.StartPod()
		}

	}

	fmt.Printf("[Pod Manager] recover with %d data", len(resList))
	return
}

func (p *PodManager) CheckIfPodExist(podName string) bool {
	_, ok := p.name2pod[podName]
	return ok
}

func (p *PodManager) DeletePod(podName string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.CheckIfPodExist(podName) {
		//不存在该pod
		return errors.New(podName + "对应的pod不存在")
	}
	pod, _ := p.name2pod[podName]
	fmt.Printf("[Kubelet] Prepare delete pod...")
	pod.DeletePod()
	delete(p.name2pod, podName)
	return nil
}

func (p *PodManager) AddPod(config *object.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	//首先检查name对应的pod是否存在， 存在的话报错
	if p.CheckIfPodExist(config.Metadata.Name) {
		return errors.New(config.Metadata.Name + "对应的pod已经存在，请先删除原pod")
	}
	newPod := pod.NewPodfromConfig(config, p.clientConfig)
	p.name2pod[config.Name] = newPod
	return nil
}

// CopyName2pod only copy the pointers in map, check before actual use
func (p *PodManager) CopyName2pod() map[string]*pod.Pod {
	p.lock.Lock()
	defer p.lock.Unlock()
	uuidMap := make(map[string]*pod.Pod)
	for key, val := range p.name2pod {
		uuidMap[key] = val
	}
	return uuidMap
}

func list(key string) ([]etcdstorage.ListRes, error) {
	fmt.Printf("[list watcher] list %s \n", key)
	resourceURL := _const.BASE_URI + key
	request, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}
	//向api-server发送请求
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var resList []etcdstorage.ListRes
	err = json.Unmarshal(data, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}
