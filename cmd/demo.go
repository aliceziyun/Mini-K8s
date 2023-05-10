package main

import (
	"Mini-K8s/pkg/etcdstorage"
	o "Mini-K8s/pkg/object"
	"Mini-K8s/pkg/tcp"
	"encoding/json"
	"fmt"
	"time"
)

func mainForTcp() {
	if false {
		//go udp.Server()
		time.Sleep(500 * time.Millisecond)
		//go udp.Client()
	} else {
		go tcp.Server("127.0.0.1:8080")
		time.Sleep(500 * time.Millisecond)
		go tcp.Client("127.0.0.1:8080", "Client 1")
	}
	time.Sleep(3 * time.Second)
}

func mainForEtcd() {

	store, err := etcdstorage.InitKVStore([]string{"127.0.0.1:2379"}, time.Second)
	if err != nil {
		return
	}

	container := o.Container{
		Name:  "nginx",
		Image: "nginx:stable-alpine",
	}

	pod := o.Pod{}

	pod.Kind = "Pod"
	pod.ApiVersion = 1
	pod.Metadata.Name = "pod-example"
	pod.Metadata.Namespace = "default"

	pod.Spec.Containers = append(pod.Spec.Containers, container)

	pod.Status.Phase = "Running"

	fmt.Println(pod)

	jsonBytes, err0 := json.Marshal(pod)
	if err0 != nil {
		return
	}

	key := etcdstorage.EtcdPodPrefix + pod.Metadata.Namespace + "/" + pod.Metadata.Name

	go func() {
		time.Sleep(1 * time.Second)

		err1 := store.Put(key, string(jsonBytes))
		if err1 != nil {
			return
		}

		value, err2 := store.Get(key)
		if err2 != nil {
			return
		}

		err3 := store.Del(key)
		if err3 != nil {
			return
		}

		podPtr := &o.Pod{}
		err4 := json.Unmarshal([]byte(value), podPtr)
		if err4 != nil {
			return
		}

		fmt.Println(*podPtr)
	}()

	go func() {
		store.Watch(key)
	}()

	time.Sleep(5 * time.Second)

}
