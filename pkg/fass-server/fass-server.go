package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/file"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type FassServer struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
	queue       queue.ConcurrentQueue
	hashMap     *_map.ConcurrentMap
	mtx         sync.Mutex
}

func NewServer(lsConfig *listwatcher.Config) *FassServer {
	println("[FassServer] fassServer create")

	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[FassServer] list watch start fail...")
	}

	server := &FassServer{
		ls:      ls,
		hashMap: _map.NewConcurrentMap(),
	}
	server.stopChannel = make(chan struct{})
	return server
}

func (s *FassServer) Run() {
	fmt.Printf("[FassServer] start running\n")
	go s.register()
	go s.worker()
	select {}
}

func (s *FassServer) register() {
	err := s.ls.Watch(_const.FUNC_RUNTIME_PREFIX, s.watchNewFunc, s.stopChannel)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[FassServer] listen fail...\n")
	}
}

func (s *FassServer) worker() {
	fmt.Printf("[FassServer] Starting...\n")
	for {
		if !s.queue.Empty() {
			metaMap := s.queue.Front()
			s.queue.Dequeue()
			s.mtx.Lock()
			err := s.invokeFunction(metaMap.(map[string]any))
			s.mtx.Unlock()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (s *FassServer) watchNewFunc(res etcdstorage.WatchRes) {
	metaMap := make(map[string]any, 10)
	err := json.Unmarshal(res.ValueBytes, &metaMap)
	if err != nil {
		fmt.Println(err)
	}

	s.queue.Enqueue(metaMap)
}

func (s *FassServer) invokeFunction(metaMap map[string]any) error {
	//从etcd中读取出Function实体
	name := fmt.Sprintln(metaMap["name"])
	name = strings.Replace(name, "\n", "", -1)
	resList, err := s.ls.List(_const.FUNC_CONFIG_PREFIX + "/" + name)
	if err != nil {
		return err
	}

	if len(resList) == 0 {
		return nil
	}
	function := &object.Function{}
	err = json.Unmarshal(resList[0].ValueBytes, function)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//读取函数
	zip, err := os.ReadFile(function.Path)
	functionBody := string(zip)

	uuid, err := uuid2.NewUUID()

	body := getFunctionBody(functionBody, function.ArgNum, function.Return, function.FuncName, uuid.String())

	//将文件放入共享目录中
	fileName := name + "-" + uuid.String() + ".py"
	dirName := path.Join(_const.SHARED_DATA_DIR, name+"-"+uuid.String())
	err = file.Bytes2File(body, fileName, dirName)

	//创建Pod
	err = invokePod(uuid.String(), fileName, dirName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func invokePod(uuid string, fileName string, dirName string) error {
	pod := object.Pod{}
	pod.Metadata.Name = fmt.Sprintf("Func-%s", uuid)
	pod.Metadata.Uid = uuid
	pod.Name = pod.Metadata.Name
	pod.Kind = object.POD

	container := object.Container{}
	container.Image = "testpy:latest"
	container.Name = fmt.Sprintf("Func-%s", uuid)
	//commands := []string{
	//	"python",
	//	fileName,
	//}
	commands := []string{
		"/bin/sh",
		"-c",
		"while true; do echo hello world; sleep 1; done",
	}
	container.Command = commands
	container.Args = nil
	volumeMounts := []object.VolumeMount{
		{
			Name:      "Serveless",
			MountPath: "/home/test",
		},
	}
	container.Ports = []object.ContainerPort{
		{Port: "6666"},
	}
	container.VolumeMounts = volumeMounts
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	volumes := []object.Volume{
		{
			Name: "Serveless",
			Type: "hostPath",
			Path: dirName,
		},
	}
	pod.Spec.Volumes = volumes

	go func() {
		fmt.Println("[FassServer] new function added!")
		err := addFunctionPod(&pod)
		if err != nil {
			fmt.Println("[FassServer]", err)
			return
		}

	}()

	return nil
}
