package pod

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/kubelet/dockerClient"
	"Mini-K8s/pkg/kubelet/message"
	"Mini-K8s/pkg/kubelet/podWorker"
	"Mini-K8s/pkg/object"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	uuid2 "github.com/google/uuid"

	"github.com/docker/docker/api/types"
)

// yaml文件中表示emptyDir和hostPath的
const emptyDir = "emptyDir"
const hostPath = "hostPath"

// pod的状态
const POD_PENDING_STATUS = "Pending"
const POD_FAILED_STATUS = "Failed"
const POD_RUNNING_STATUS = "Running"
const POD_EXITED_STATUS = "Exited"
const POD_DELETED_STATUS = "deleted"
const POD_CREATEED_STATUS = "created"

// container的状态
const CONTAINER_EXITED_STATUS = "exited"
const CONTAINER_RUNNING_STATUS = "running"
const CONTAINER_CREATED_STATUS = "created"

// pod的liveness probe间隔,为了防止探针command拥堵，要等上一次的response
const PROBE_INTERVAL = 60 //探针间隔=60s

type Pod struct {
	configPod   *object.Pod
	containers  []object.ContainerMeta
	tmpDirMap   map[string]string
	hostDirMap  map[string]string
	hostFileMap map[string]string
	//读写锁
	rwLock       sync.RWMutex
	commandChan  chan message.PodCommand
	responseChan chan message.PodResponse
	podWorker    *podWorker.PodWorker
	//探针相关
	timer        *time.Ticker
	canProbeWork bool //控制是否能probe；每次probe后要暂设为false
	stopChan     chan bool
}

type PodNetWork struct {
	OpenPortSet []string //开放端口集合
	GateWay     string   //网关地址 ip4
	Ipaddress   string   //在docker网段中的地址

}

func (p *Pod) GetName() string {
	return p.configPod.Name
}

func (p *Pod) GetLabel() map[string]string {
	// return p.configPod.Labels
	return nil
}
func (p *Pod) GetUid() string {
	// return p.configPod.UID
	return ""
}

func (p *Pod) GetContainers() []object.ContainerMeta {
	fmt.Println("get containers")
	deepContainers := p.containers
	return deepContainers
}

// 修改pod的status，如果不用修改则返回falses
func (p *Pod) compareAndSetStatus(status string) bool {
	oldStatus := p.getStatus()
	if oldStatus == status {
		return false
	}
	p.configPod.Status.Phase = status
	return true
}
func (p *Pod) getStatus() string {
	return p.configPod.Status.Phase
}

// Pod: 更新pod
func (p *Pod) uploadPod() {
	err := updatePod(p.configPod)
	if err != nil {
		fmt.Println("[pod] updateRuntimePod error" + err.Error())
	}
}

// 初始化pod
func NewPodfromConfig(config *object.Pod, clientConfig client.Config) *Pod {
	newPod := &Pod{}
	newPodMeta := &object.PodMeta{}
	newPod.configPod = config
	// newPod.configPod.Ctime = time.Now().Format("2020-06-01 15:40:00")
	newPod.canProbeWork = false
	var rwLock sync.RWMutex
	newPod.rwLock = rwLock
	newPod.commandChan = make(chan message.PodCommand, 100)
	newPod.responseChan = make(chan message.PodResponse, 100)
	newPod.podWorker = &podWorker.PodWorker{}
	////创建pod里的containers同时把config里的originName替换为realName
	//先填第一个pause容器
	newPod.containers = append(newPod.containers, object.ContainerMeta{
		OriginName: "pause",
		RealName:   "", //先设置为空
	})
	pauseRealName := "pause"
	for index, value := range config.Spec.Containers {
		//realName := config.Name + "_" + value.Name
		realName := config.Name + value.Name
		newPod.containers = append(newPod.containers, object.ContainerMeta{
			OriginName: value.Name,
			RealName:   realName,
		})
		pauseRealName += "_" + realName
		config.Spec.Containers[index].Name = realName
	}
	newPod.containers[0].RealName = pauseRealName
	err := newPod.AddVolumes(config.Spec.Volumes)
	if err != nil {
		newPod.compareAndSetStatus(POD_FAILED_STATUS)
	} else {
		newPod.compareAndSetStatus(POD_PENDING_STATUS)
	}
	//启动pod
	newPod.StartPod()
	//生成创建pod容器的command
	commandWithConfig := &message.CommandWithConfig{}
	commandWithConfig.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	commandWithConfig.Group = config.Spec.Containers
	// 把config中容器的volumeMounts MountPath 换成实际路径
	for _, value := range commandWithConfig.Group {
		if value.VolumeMounts != nil {
			for index, it := range value.VolumeMounts {
				path, ok := newPod.tmpDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPod.hostDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPod.hostFileMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				fmt.Println("[Kubelet] error:container Mount path didn't exist")
			}
		}
	}
	podCommand := message.PodCommand{
		ContainerCommand: &(commandWithConfig.Command),
		PodCommandType:   message.ADD_POD,
	}
	//更新并上传podMeta
	newPodMeta.PodName = newPod.configPod.Name
	newPodMeta.NodeName = _const.NODE_NAME
	newPodMeta.Containers = newPod.containers
	newPodMeta.HostDirMap = newPod.hostDirMap
	newPodMeta.HostFileMap = newPod.hostFileMap
	newPodMeta.TmpDirMap = newPod.tmpDirMap
	err = uploadPodMeta(newPodMeta)
	if err != nil {
		return nil
	}

	newPod.commandChan <- podCommand
	return newPod
}

func (p *Pod) StartPod() {
	go p.podWorker.SyncLoop(p.commandChan, p.responseChan) //每个Pod有一个对应的worker
	go p.listeningResponse()
	p.canProbeWork = true
	p.StartProbe()
}

func (p *Pod) listeningResponse() {
	//删除pod后释放资源
	defer p.releaseResource()
	for {
		select {
		case response, ok := <-p.responseChan:
			if !ok {
				return
			}
			switch response.PodResponseType {
			case message.ADD_POD:
				p.rwLock.Lock()
				responseWithContainIds := (*message.ResponseWithContainIds)(unsafe.Pointer(response.ContainerResponse))
				fmt.Printf("[pod] receive AddPod responce")
				fmt.Println(*responseWithContainIds)
				fmt.Println(*responseWithContainIds.NetWorkInfos)
				//先看response是否是操作成功了
				if responseWithContainIds.Err != nil {
					//操作出错了
					if p.SetStatusAndErr(POD_FAILED_STATUS, responseWithContainIds.Err) {
						p.SetContainersAndStatus(responseWithContainIds.Containers, POD_RUNNING_STATUS)
						p.setIpAddress(responseWithContainIds.NetWorkInfos)
						p.uploadPod()
					}
					fmt.Println(responseWithContainIds.Err.Error())
				} else {
					//成功添加pod，将其状态变成running
					p.SetContainersAndStatus(responseWithContainIds.Containers, POD_RUNNING_STATUS)
					p.setIpAddress(responseWithContainIds.NetWorkInfos)
					p.uploadPod()
				}
				p.rwLock.Unlock()
			case message.PROBE_POD:
				p.rwLock.Lock()
				if p.getStatus() == POD_DELETED_STATUS {
					p.canProbeWork = false
					p.rwLock.Unlock()
				} else {
					responseWithProbeInfos := (*message.ResponseWithProbeInfos)(unsafe.Pointer(response.ContainerResponse))
					if responseWithProbeInfos.Err != nil {
						p.SetStatusAndErr(POD_FAILED_STATUS, responseWithProbeInfos.Err)
					} else {
						status := POD_RUNNING_STATUS
						for _, value := range responseWithProbeInfos.ProbeInfos {
							if value == CONTAINER_CREATED_STATUS {
								status = POD_CREATEED_STATUS
								break
							}
							if value == CONTAINER_EXITED_STATUS {
								status = POD_EXITED_STATUS
								cli, _ := dockerClient.GetNewClient()
								for _, val := range p.containers {
									_ = cli.ContainerStop(context.Background(), val.ContainerId, nil)
								}
								break
							}
						}
						if p.compareAndSetStatus(status) {
							p.uploadPod()
						}
					}
					p.canProbeWork = true
					p.rwLock.Unlock()
				}

			case message.DELETE_POD:
				//
				return
			}
		}
	}
}

func (p *Pod) ReceivePodCommand(podCommand message.PodCommand) {
	p.commandChan <- podCommand
}

func (p *Pod) AddVolumes(volumes []object.Volume) error {
	p.tmpDirMap = make(map[string]string)
	p.hostDirMap = make(map[string]string)
	p.hostFileMap = make(map[string]string)
	for _, value := range volumes {
		if value.Type == emptyDir {
			//临时目录，随机生成
			u, _ := uuid2.NewUUID()
			path := GetCurrentAbPathByCaller() + "/tmp/" + u.String()
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return err
			}
			p.tmpDirMap[value.Name] = path
		} else if value.Type == hostPath {
			//指定了实际目录
			_, err := os.Stat(value.Path)
			if err != nil {
				err := os.MkdirAll(value.Path, os.ModePerm)
				if err != nil {
					return err
				}
			}
			p.hostDirMap[value.Name] = value.Path
		} else {
			//文件映射
			_, err := os.Stat(value.Path)
			if err != nil {
				return err
			}
			p.hostFileMap[value.Name] = value.Path
		}
	}
	return nil
}

// 获取当前文件的路径，
func GetCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

// --------------------辅助函数-------------------------------- //
func (p *Pod) SetStatusAndErr(status string, err error) bool {
	// p.configPod.Status.Err = err.Error()
	return p.compareAndSetStatus(status)
}
func (p *Pod) SetContainersAndStatus(containers []object.ContainerMeta, status string) bool {
	for _, value := range containers {
		for index, it := range p.containers {
			if it.RealName == value.RealName {
				p.containers[index].ContainerId = value.ContainerId
			}
		}
	}
	return p.compareAndSetStatus(status)
}
func (p *Pod) setIpAddress(settings *types.NetworkSettings) {
	p.configPod.Status.PodIP = settings.IPAddress
}
func filterSingle(input string) string {
	index := strings.Index(input, "/tcp")
	return input[0:index]
}
func filterChars(input []string) []string {
	var result []string
	for _, value := range input {
		result = append(result, filterSingle(value))
	}
	return result
}

func (p *Pod) StartProbe() {
	p.timer = time.NewTicker(PROBE_INTERVAL * time.Second)
	p.stopChan = make(chan bool)
	go func(p *Pod) {
		defer p.timer.Stop()
		for {
			select {
			case <-p.timer.C:
				p.rwLock.Lock()
				// 当前状态为running且探针可用时监测
				if p.canProbeWork && p.getStatus() != POD_PENDING_STATUS && p.getStatus() != POD_FAILED_STATUS && p.getStatus() != POD_DELETED_STATUS && p.getStatus() != POD_EXITED_STATUS {
					command := &message.CommandWithContainerIds{}
					command.CommandType = message.COMMAND_PROBE_CONTAINER
					var group []string
					for _, value := range p.containers {
						group = append(group, value.ContainerId)
					}
					command.ContainerIds = group
					podCommand := message.PodCommand{
						PodCommandType:   message.PROBE_POD,
						ContainerCommand: &(command.Command),
					}
					p.commandChan <- podCommand
					p.canProbeWork = false
				}
				p.rwLock.Unlock()
			case stop := <-p.stopChan:
				if stop {
					return
				}
			}
		}
	}(p) //p作为参数，调用这个函数
}

func (p *Pod) DeletePod() {
	p.rwLock.Lock()
	fmt.Println("[Kubelet] into deletePod")
	p.compareAndSetStatus(POD_DELETED_STATUS)
	command := &message.CommandWithContainerIds{}
	command.CommandType = message.COMMAND_DELETE_CONTAINER
	var group []string
	for _, value := range p.containers {
		group = append(group, value.ContainerId)
	}
	command.ContainerIds = group
	podCommand := message.PodCommand{
		PodCommandType:   message.DELETE_POD,
		ContainerCommand: &(command.Command),
	}
	p.commandChan <- podCommand
	fmt.Println("[Kubelet] send command")
	deleteRuntimePod(p.GetName())
	p.rwLock.Unlock()
}

func RecoverPod(meta *object.PodMeta, configPod *object.Pod) *Pod {
	var rwLock sync.RWMutex
	pod := &Pod{}
	pod.configPod = configPod
	pod.hostFileMap = meta.HostFileMap
	pod.hostDirMap = meta.HostDirMap
	pod.tmpDirMap = meta.TmpDirMap
	pod.containers = meta.Containers
	pod.podWorker = &podWorker.PodWorker{}
	pod.canProbeWork = false
	pod.rwLock = rwLock
	pod.commandChan = make(chan message.PodCommand, 100)
	pod.responseChan = make(chan message.PodResponse, 100)

	return pod
}

func uploadPodMeta(meta *object.PodMeta) error {
	metaRaw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	reqBody := bytes.NewBuffer(metaRaw)

	req, err := http.NewRequest("PUT", _const.BASE_URI+_const.POD_META_PREFIX+"/"+meta.PodName, reqBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("status code not 200")
	}
	return nil
}

func (p *Pod) releaseResource() {
	p.rwLock.Lock()
	p.canProbeWork = false
	p.stopChan <- true
	close(p.commandChan)
	close(p.responseChan)
	p.rwLock.Unlock()
}
