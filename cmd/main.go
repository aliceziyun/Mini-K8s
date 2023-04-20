package main

import (
	"Mini-K8s/pkg/etcdstorage"
	"time"
)

func main() {

	store, err := etcdstorage.InitKVStore([]string{"127.0.0.1:2379"}, time.Second)
	if err != nil {
		return
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		store.Put("/etcd/1", "v1")
		store.Put("/etcd/1", "v2")
		store.Del("/etcd/1")
	}()
	store.Watch("/etcd/1")

	//store.Put("/etcd2/1", "value1")
	//store.Put("/etcd2/2", "value2")
	//store.Put("/etcd2/3", "value3")
	//
	//store.Get("/etcd2/1")
	//
	//store.Del("/etcd2/1")
	//
	//store.GetPrefix("/etcd2")

}
