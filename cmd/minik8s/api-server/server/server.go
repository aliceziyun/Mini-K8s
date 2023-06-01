package apiserver

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
	_map "Mini-K8s/third_party/map"
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

var counter = 1
var mtx sync.Mutex

type APIServer struct {
	engine       *gin.Engine
	port         int
	store        *etcdstorage.KVStore
	publisher    *message.Publisher
	watcherMtx   sync.Mutex
	watcherChan  chan watchOpt
	ticketSeller *atomic.Uint64
	recordTable  *_map.ConcurrentMap
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
		engine:       engine,
		port:         c.HttpPort,
		store:        store,
		publisher:    publisher,
		watcherChan:  watcherChan,
		ticketSeller: atomic.NewUint64(0),
		recordTable:  _map.NewConcurrentMap(),
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
	//-------------------RESTful------------------------
	engine.POST(_const.PATH, s.watch)
	engine.PUT(_const.PATH, s.put)
	engine.DELETE(_const.PATH, s.delete)
	engine.GET(_const.PATH, s.get)

	engine.POST(_const.PATH_PREFIX, s.watch)
	engine.GET(_const.PATH_PREFIX, s.getByPrefix)

	engine.PUT(_const.POD_CONFIG, s.addPodConfig)
	engine.GET(_const.POD_CONFIG_PREFIX, s.getByPrefix)
	engine.DELETE(_const.POD_CONFIG_PREFIX, s.deletePod)

	engine.PUT(_const.POD_RUNTIME_PREFIX, s.addPodRuntime)

	engine.PUT(_const.RS_CONFIG, s.addRS)
	engine.GET(_const.RS_CONFIG, s.get)

	engine.DELETE(_const.RS_CONFIG_PREFIX, s.deleteRS)

	engine.PUT(_const.SERVICE_CONFIG, s.addService)

	engine.PUT(_const.DNS_CONFIG, s.addDNS)

	engine.PUT(_const.NODE_CONFIG, s.addNode)

	//-------------------non-REST------------------------
	engine.POST(_const.SERVELESS_PATH, s.invoke)
	engine.POST(_const.SERVERLESS_CALLBACK_PATH, s.receive)
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
			//if opt.withPrefix {
			//	_, resChan = s.store.WatchWithPrefix(key)
			//} else {
			_, resChan = s.store.Watch(key)
			//}

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
