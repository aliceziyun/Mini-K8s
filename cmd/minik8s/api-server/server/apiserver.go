package apiserver

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"sync"
)

type watchOpt struct {
	key        string
	withPrefix bool
	ticket     uint64
}

type APIServer struct {
	engine *gin.Engine
	port   int
	//resourceSet  mapset.Set[string]
	store     *etcdstorage.KVStore
	publisher *message.Publisher
	//watcherMap   map[string]*watcher
	watcherMtx   sync.Mutex // watcherMtx 保护watcherCount
	watcherChan  chan watchOpt
	ticketSeller *atomic.Uint64
}

func NewServer(c *ServerConfig) (*APIServer, error) {
	engine := gin.Default()
	store, err := etcdstorage.InitKVStore(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		fmt.Println("Error connecting to etcd.")
		fmt.Println(err.Error())
		return nil, err
	}
	publisher, err := message.NewPublisher(c.QueueConfig)
	if err != nil {
		fmt.Println("Error connecting to rabbitmq.")
		fmt.Println(err.Error())
		return nil, err
	}
	watcherChan := make(chan watchOpt)
	//kubeNetSupport, err2 := kubeNetSupport.NewKubeNetSupport(listerwatcher.DefaultConfig(), client.DefaultClientConfig())
	//if err2 != nil {
	//	return nil, err2
	//}
	s := &APIServer{
		engine: engine,
		port:   c.HttpPort,
		//resourceSet:  mapset.NewSet[string](c.ValidResources...),
		store:     store,
		publisher: publisher,
		//watcherMap:   map[string]*watcher{},
		watcherChan:  watcherChan,
		ticketSeller: atomic.NewUint64(0),
		//kubeNetSupport: kubeNetSupport,
	}

	engine.PUT("/testpod", s.addPodTest)

	go s.daemon(watcherChan)

	return s, nil
}

func (s *APIServer) Run() error {
	// start web api
	err := s.engine.Run(fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	return err
}

func (s *APIServer) addPodTest(ctx *gin.Context) {
	key := "1"
	pod := object.Pod{}
	fmt.Printf("key:%v\n", key)

	pod.Kind = "Pod"
	pod.ApiVersion = 1
	pod.Metadata.Name = "pod-example"
	pod.Metadata.Namespace = "default"

	pod.Status.Phase = "Running"

	body, _ := json.Marshal(pod)

	err := s.store.Put(key, string(body))
	err = s.store.Del(key)
	if err != nil {
		return
	}
}

func (s *APIServer) daemon(listening <-chan watchOpt) {
	fmt.Println("start daemon")
	var resChan <-chan etcdstorage.WatchRes
	s.store.Watch("1")
	go func(resChan <-chan etcdstorage.WatchRes) {
		for res := range resChan {
			data, err := json.Marshal(res)
			if err != nil {
				fmt.Println("error when watch etcd")
			}
			err = s.publisher.Publish("1", data, "application/json")
			if err != nil {
				fmt.Println("error when publish")
			}
		}
	}(resChan)
}
