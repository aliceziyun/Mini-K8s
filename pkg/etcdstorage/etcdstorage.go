package etcdstorage

import (
	"context"
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
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Put(context.TODO(), key, val)
	return err
}

func (kvs *KVStore) Del(key string) error {
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Delete(context.TODO(), key)
	return err
}

func (kvs *KVStore) Watch(key string) {
	kv := clientv3.NewKV(kvs.client)

	getResp, err := kv.Get(context.TODO(), key)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 当前etcd集群事务ID, 单调递增的
	watchStartRevision := getResp.Header.Revision + 1

	watcher := clientv3.NewWatcher(kvs.client)

	// 创建一个 5s 后取消的上下文
	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	// 该监听动作在 5s 后取消
	watchRespChan := watcher.Watch(ctx, key, clientv3.WithRev(watchStartRevision))

	// 处理kv变化事件
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			fmt.Print("[WATCH]")
			switch event.Type {
			case mvccpb.PUT:
				//fmt.Println("Put: ", string(event.Kv.Value), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
				fmt.Println("Put\tRevision: ", event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("Delete\tRevision:", event.Kv.ModRevision)
			}
		}
	}
}
