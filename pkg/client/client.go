package client

import (
	"Mini-K8s/pkg/message/config"
	"time"
)

type Config struct {
	Host string // ip and port
}

// 比较RESTful的
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

type QueryRes struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	ResultType  string   `json:"resultType"`
	ResultArray []Result `json:"result"`
}

type Result struct {
	Value []interface{} `json:"value"`
}
