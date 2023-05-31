package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/fass-server/fass-util"
	workflow "Mini-K8s/pkg/fass-server/workflow"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"encoding/json"
	"fmt"
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
	resChannel  chan string
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
	server.resChannel = make(chan string)
	return server
}

func (s *FassServer) Run() {
	fmt.Printf("[FassServer] start running\n")
	go s.register()
	go s.worker()
	select {}
}

func (s *FassServer) register() {
	go func() {
		err := s.ls.Watch(_const.FUNC_RUNTIME_PREFIX, s.watchNewFunc, s.stopChannel)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("[FassServer] listen fail...\n")
		}
	}()

	go func() {
		err := s.ls.Watch(_const.WORKFLOW_CONFIG_PREFIX, s.watchNewWorkflow, s.stopChannel)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("[FassServer] listen fail...\n")
		}
	}()

}

func (s *FassServer) worker() {
	fmt.Printf("[FassServer] Starting...\n")
	for {
		if !s.queue.Empty() {
			meta := s.queue.Front()
			s.queue.Dequeue()
			err := fass_util.InvokeFunction(meta.(*object.FunctionMeta), s.ls)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (s *FassServer) watchNewFunc(res etcdstorage.WatchRes) {
	//说明有任务做完了
	if res.ResType == etcdstorage.DELETE {
		fields := strings.Split(res.Key, "/")
		funcName := fields[len(fields)-1]
		fmt.Printf("[FassServer] function %s finish! Please go to the shared directory to check the result!", funcName)
		//这里可以把代码文件删了，不过也没必要
		return
	}

	meta := &object.FunctionMeta{}
	err := json.Unmarshal(res.ValueBytes, &meta)
	if err != nil {
		fmt.Println(err)
	}

	s.queue.Enqueue(meta)
}

func (s *FassServer) watchNewWorkflow(res etcdstorage.WatchRes) {
	wf := &object.WorkFlow{}
	err := json.Unmarshal(res.ValueBytes, &wf)
	if err != nil {
		fmt.Println(err)
	}
	//新建一个workflow manager，由manager管理workflow
	go workflow.NewWorkflowManager(wf, s.resChannel).Run()
}
