package apiserver

import (
	"Mini-K8s/pkg/message"
	"time"
)

type ServerConfig struct {
	HttpPort       int
	ValidResources []string // 合法的resource
	EtcdEndpoints  []string // etcd集群每一个节点的ip和端口
	EtcdTimeout    time.Duration
	QueueConfig    *message.QConfig
	Recover        bool
}

var defaultValidResources = []string{"pod"}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		HttpPort:       8080,
		ValidResources: defaultValidResources,
		EtcdEndpoints:  []string{"localhost:2379"}, //设置etcd的端口号
		EtcdTimeout:    5 * time.Second,
		QueueConfig:    message.DefaultQConfig(),
		Recover:        false,
	}
}
