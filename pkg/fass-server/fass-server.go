package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"encoding/json"
	"fmt"
	"sync"
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
	select {}
}

func (s *FassServer) register() {
	err := s.ls.Watch(_const.FUNC_RUNTIME_PREFIX, s.watchNewFunc, s.stopChannel)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[FassServer] listen fail...\n")
	}
}

func (s *FassServer) watchNewFunc(res etcdstorage.WatchRes) {
	metaMap := make(map[string]any, 10)
	err := json.Unmarshal(res.ValueBytes, &metaMap)
	if err != nil {
		fmt.Println(err)
	}

	//从etcd中读取出实体，修饰函数体，并准备创建pod
}
