package main

import (
	"Mini-K8s/pkg/etcdstorage"
	o "Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"time"
)

func main() {

	store, err0 := etcdstorage.InitKVStore([]string{"127.0.0.1:2379"}, time.Second)
	if err0 != nil {
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

		fmt.Println(value)

		err3 := store.Del(key)
		if err3 != nil {
			return
		}
	}()

	store.Watch(key)

}
