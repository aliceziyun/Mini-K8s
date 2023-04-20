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

	store.Put("/etcd2/1", "value1")
	store.Put("/etcd2/2", "value2")

	store.Get("/etcd2/1")

	store.Del("/etcd2/1")

	store.GetPrefix("/etcd2")

	//etcdStorage.EtcdWatchTest()
}
