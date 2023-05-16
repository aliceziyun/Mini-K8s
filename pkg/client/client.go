package client

import (
	"Mini-K8s/pkg/message/config"
	"time"
)

type Config struct {
	Host string // ip and port
}

type RESTClient struct {
	Base string // url = base+resource+name
}

type ServerConfig struct {
	HttpPort       int
	ValidResources []string
	EtcdEndpoints  []string
	EtcdTimeout    time.Duration
	QueueConfig    *config.QConfig
	//Recover        bool
}

func DefaultClientConfig() Config {
	return Config{
		Host: "127.0.0.1:8080",
	}
}

