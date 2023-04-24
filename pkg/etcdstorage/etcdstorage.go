package etcdstorage

import (
	"Mini-K8s/pkg/message"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"

	"github.com/coreos/etcd/clientv3"
)

const (
	EtcdPodPrefix string = "/registry/pods/"
)

type KVStore struct {
	client *clientv3.Client
}

type WatchRes struct {
	ResType         int
	ResourceVersion int64
	CreateVersion   int64
	IsCreate        bool // true when ResType == PUT and the key is new
	IsModify        bool // true when ResType == PUT and the key is old
	Key             string
	ValueBytes      []byte
}

// ref https://blog.csdn.net/wohu1104/article/details/108552649

func InitKVStore(endpoints []string, timeout time.Duration) (*KVStore, error) {
	fmt.Print("\n")
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	}

	// establish a client
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	return &KVStore{client: cli}, nil
}

func (kvs *KVStore) Get(key string) (string, error) {
	kv := clientv3.NewKV(kvs.client)
	response, err := kv.Get(context.TODO(), key)
	if err != nil {
		return "", err
	}

	if len(response.Kvs) != 0 {
		return string(response.Kvs[0].Value), nil
	} else {
		return "", nil
	}
}

func (kvs *KVStore) GetPrefix(key string) error {
	kv := clientv3.NewKV(kvs.client)
	response, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	if len(response.Kvs) != 0 {
		fmt.Print("-> Get result:\n")
		for _, resp := range response.Kvs {
			fmt.Printf("\tkey: %s, value: %s\n", string(resp.Key), string(resp.Value))
		}
	} else {
		fmt.Println("-> Get result: Empty")
	}
	fmt.Print("\n")
	return nil
}

func (kvs *KVStore) Put(key string, val string) error {
	fmt.Println("put a new pod", key, val)
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Put(context.TODO(), key, val)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (kvs *KVStore) Del(key string) error {
	fmt.Println("delete a new pod", key)
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Delete(context.TODO(), key)
	return err
}

func (kvs *KVStore) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	fmt.Println("etcd start watch", key)

	watchResChan := make(chan WatchRes)

	watcher := clientv3.NewWatcher(kvs.client)
	ctx, cancel := context.WithCancel(context.TODO())

	watch := func(c chan<- WatchRes) {
		//fmt.Println("watch again")
		watchRespChan := watcher.Watch(ctx, key)
		// 处理kv变化事件
		for watchResp := range watchRespChan {
			var res WatchRes
			for _, event := range watchResp.Events {
				fmt.Print("[WATCH]")
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("Put\tRevision: ", event.Kv.CreateRevision, event.Kv.ModRevision)
					data, _ := json.Marshal("sewgwq")
					publisher, _ := message.NewPublisher(message.DefaultQConfig())
					publisher.Publish("/testwatch", data, "application/json")
				case mvccpb.DELETE:
					fmt.Println("Delete\tRevision:", event.Kv.ModRevision)
				}
			}
			c <- res
		}
		fmt.Println("etcd close watcher with key", key)
		close(c)
	}

	go watch(watchResChan)

	return cancel, watchResChan
}
