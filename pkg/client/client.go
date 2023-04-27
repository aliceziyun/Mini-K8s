package client

import (
	"Mini-K8s/pkg/message"
	"Mini-K8s/pkg/object"
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
	QueueConfig    *message.QConfig
	//Recover        bool
}

func DefaultClientConfig() Config {
	return Config{
		Host: "127.0.0.1:8080",
	}
}

func (r RESTClient) UpdateRuntimePod(pod *object.Pod) error {
	// attachURL := "/registry/pod/default/" + pod.Name
	// err := Put(r.Base+attachURL, pod)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (r RESTClient) DeleteRuntimePod(podName string) error {
	// attachURL := "/registry/pod/default/" + podName
	// err := Del(r.Base + attachURL)
	// return err
	return nil
}

func (r RESTClient) UpdateConfigPod(pod *object.Pod) error {
	// attachURL := config.PodConfigPREFIX + "/" + pod.Name
	// err := Put(r.Base+attachURL, pod)
	// return err
	return nil
}
func (r RESTClient) DeleteConfigPod(podName string) error {
	// attachURL := config.PodConfigPREFIX + "/" + podName
	// err := Del(r.Base + attachURL)
	// return err
	return nil
}
