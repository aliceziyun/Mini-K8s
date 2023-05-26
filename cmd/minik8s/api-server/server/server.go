package apiserver

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
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

// NewServer :
// 1.连接etcd和amqp
// 2.创建新API-server
// 3.注册函数
// 4.监听
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

	registerWebFunc(engine, s)

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

// registerWebFunc: 将API方法名和方法绑定
func registerWebFunc(engine *gin.Engine, s *APIServer) {
	engine.PUT("/testAddPod", s.addPodTest)
	engine.POST("/testwatch", s.watch)
}

func (s *APIServer) daemon(listening <-chan watchOpt) {
	fmt.Println("start daemon")
	//var resChan = make(<-chan etcdstorage.WatchRes)

	//data, _ := json.Marshal("sewgwq")
	_, _ = s.store.Watch("test")
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
