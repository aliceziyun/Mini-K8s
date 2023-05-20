package apiserver

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
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
	engine.POST(_const.PATH, s.watch)
	engine.PUT(_const.POD_CONFIG, s.addPod)
	engine.GET(_const.POD_CONFIG_PREFIX, s.get)

	engine.PUT(_const.RS_CONFIG, s.addRS)
	engine.GET(_const.RS_CONFIG, s.get)

	engine.GET(_const.RS_CONFIG_PREFIX, s.get)

}

func (s *APIServer) daemon(listening <-chan watchOpt) {
	fmt.Println("[API-Server] start daemon...")
	for {
		select {
		case opt := <-listening:
			fmt.Println("[API-Server] receive new watch Opt...")
			key := opt.key

			s.watcherMtx.Lock()
			var resChan <-chan etcdstorage.WatchRes
			_, resChan = s.store.Watch(key)

			go func(resChan <-chan etcdstorage.WatchRes) {
				for res := range resChan {
					fmt.Printf("[API-Server] publish %s... \n", key)
					data, err := json.Marshal(res)
					err = s.publisher.Publish(key, data, "application/json")
					if err != nil {
						fmt.Printf("[API-Server] publish %s fail \n", key)
					}
				}
			}(resChan)

			s.watcherMtx.Unlock()
		}
	}
}
