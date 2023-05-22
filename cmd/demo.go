package main

import (
	"Mini-K8s/pkg/kubeproxy"
	"Mini-K8s/pkg/listwatcher"
	o "Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	kubeproxy.TestDns()
}

func nmain() {

	kubeProxy := kubeproxy.NewKubeProxy(listwatcher.DefaultConfig())
	kubeProxy.Run()

	time.Sleep(1 * time.Second)

	pnip1 := o.PodNameAndIp{
		Name: "pod1",
		Ip:   "192.168.1.15",
	}
	pnip2 := o.PodNameAndIp{
		Name: "pod2",
		Ip:   "192.168.1.6",
	}

	servPort := o.ServicePort{
		Name:       "sp",
		Port:       "8000",
		TargetPort: "8081",
	}

	serviceSpec := o.ServiceSpec{
		Type:      "ClusterIp",
		ClusterIp: "10.10.0.5",
	}
	serviceSpec.PodNameAndIps = append(serviceSpec.PodNameAndIps, pnip1, pnip2)
	serviceSpec.Ports = append(serviceSpec.Ports, servPort)

	service := o.Service{
		ApiVersion: 1,
		Kind:       "SERVICE",
	}
	service.Metadata.Name = "service1"
	service.Metadata.Namespace = "default"

	service.Spec = serviceSpec

	//fmt.Println(service)

	jsonBytes, err0 := json.Marshal(service)
	if err0 != nil {
		return
	}

	// todo put
	fmt.Println(jsonBytes)
}
