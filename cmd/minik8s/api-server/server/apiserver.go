package apiserver

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"net/http"
	"strconv"
	"sync"
)

type watchOpt struct {
	key        string
	withPrefix bool
	ticket     uint64
}

type Ticket struct {
	T uint64
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
	engine.POST("/testwatch", s.watch)

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
	key := "2"
	pod := object.Pod{}
	fmt.Printf("key:%v\n", key)

	pod.Kind = "Pod"
	pod.ApiVersion = 1
	pod.Metadata.Name = "pod-example"
	pod.Metadata.Namespace = "default"

	pod.Status.Phase = "Running"

	body, _ := json.Marshal(pod)

	err := s.store.Put(key, string(body))
	if err != nil {
		return
	}
}

func (s *APIServer) daemon(listening <-chan watchOpt) {
	fmt.Println("start daemon")
	//var resChan = make(<-chan etcdstorage.WatchRes)

	//data, _ := json.Marshal("sewgwq")
	_, _ = s.store.Watch("2")
	//err := s.publisher.Publish("/testwatch", data, "application/json")

	//go func(resChan <-chan etcdstorage.WatchRes) {
	//	for res := range resChan {
	//		fmt.Println("watched something")
	//		data, err := json.Marshal(res)
	//		if err != nil {
	//			fmt.Println("error when watch etcd")
	//		}
	//		err = s.publisher.Publish("1", data, "application/json")
	//		if err != nil {
	//			fmt.Println("error when publish")
	//		}
	//	}
	//}(resChan)
}

func (s *APIServer) watch(ctx *gin.Context) {
	//key := ctx.Request.URL.Path
	ticketStr, status := ctx.GetPostForm("ticket")
	fmt.Println(ticketStr, status)
	if !status {
		t := Ticket{}
		t.T = s.ticketSeller.Add(1)
		data, _ := json.Marshal(t)
		//s.watcherChan <- watchOpt{key: key, withPrefix: false, ticket: t.T}
		ctx.Data(http.StatusOK, "application/json", data)
	} else {
		s.watcherMtx.Lock()
		ticket, err := strconv.ParseUint(ticketStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			ctx.AbortWithStatus(http.StatusBadRequest)
		} else {
			//if s.watcherMap[key] != nil {
			//	s.watcherMap[key].set.Remove(ticket)
			//	if s.watcherMap[key].set.Equal(mapset.NewSet[uint64]()) {
			//		s.watcherMap[key].cancel()
			//		s.watcherMap[key] = nil
			//		klog.Infof("Cancel the watcher of key %s\n", key)
			//	}
			//}
			fmt.Println(ticket, "ok")
			ctx.Status(http.StatusOK)
		}
		s.watcherMtx.Unlock()
	}
}
