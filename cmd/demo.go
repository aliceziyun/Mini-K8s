package main

import (
	"Mini-K8s/pkg/kubeproxy"
	o "Mini-K8s/pkg/object"
)

func main() {
	kubeproxy.TestDns()
}

func tmain() {

	kubeProxy := kubeproxy.NewKubeProxy()
	kubeProxy.Run()

	pnip1 := o.PodNameAndIp{
		Name: "pod1",
		Ip:   "192.168.1.100",
	}
	pnip2 := o.PodNameAndIp{
		Name: "pod2",
		Ip:   "192.168.1.101",
	}

	servPort := o.ServicePort{
		Name:       "sp",
		Port:       "8000",
		TargetPort: "8080",
	}

	serviceSpec := o.ServiceSpec{
		Type:      "ClusterIp",
		ClusterIp: "10.10.0.1",
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

	kubeproxy.TestService(service)

	//fmt.Println(service)

	//jsonBytes, err0 := json.Marshal(service)
	//if err0 != nil {
	//	return
	//}
	//
	//store, err := etcdstorage.InitKVStore([]string{"127.0.0.1:2379"}, time.Second)
	//if err != nil {
	//	return
	//}
	//key := etcdstorage.EtcdServicePrefix + service.Metadata.Namespace + "/" + service.Metadata.Name
	//
	//servPtr := &o.Service{}
	//err1 := store.Put(key, string(jsonBytes))
	//if err1 != nil {
	//	return
	//}
	//
	//go func() {
	//	store.Watch(key)
	//}()
	//time.Sleep(1 * time.Second)
	//
	//value, err2 := store.Get(key)
	//if err2 != nil {
	//	return
	//}
	//
	//err3 := store.Del(key)
	//if err3 != nil {
	//	return
	//}
	//
	//err4 := json.Unmarshal([]byte(value), servPtr)
	//if err4 != nil {
	//	return
	//}
	//
	//fmt.Println(*servPtr)

}
