package etcdstorage

import (
	"Mini-K8s/pkg/message"
	"Mini-K8s/pkg/message/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const (
	PUT    int = 0
	DELETE int = 1
)

type KVStore struct {
	client *clientv3.Client
}

type ListRes struct {
	ResourceVersion int64
	CreateVersion   int64
	Key             string
	ValueBytes      []byte
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

func InitKVStore(endpoints []string, timeout time.Duration) (*KVStore, error) {
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
	return nil
}

func (kvs *KVStore) Put(key string, val string) error {
<<<<<<< HEAD
=======
	fmt.Println("[ETCD] PUT\n", key, val)
>>>>>>> 22c9598c726809453582ec62946941ca15843c80
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Put(context.TODO(), key, val)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("put a new pod", key, val)
	return err
}

func (kvs *KVStore) Del(key string) error {
	fmt.Println("[ETCD] DELETE\n", key)
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Delete(context.TODO(), key)
	return err
}

func (kvs *KVStore) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	fmt.Println("[ETCD] WATCH\n", key)

	watchResChan := make(chan WatchRes)

	watcher := clientv3.NewWatcher(kvs.client)
	ctx, cancel := context.WithCancel(context.TODO())

<<<<<<< HEAD
	val, _ := kvs.client.Get(ctx, key)

	watchStartRevision := val.Header.Revision + 1 //获取revision,观察这个revision之后的变化

	//go watch(watchResChan)
	//time.AfterFunc(10*time.Second, func() {
	//	cancel()
	//},
	//)
	watchRespChan := watcher.Watch(ctx, key, clientv3.WithRev(watchStartRevision))

	// 处理kv变化事件
	for watchResp := range watchRespChan {
		var res WatchRes
		for _, event := range watchResp.Events {
			fmt.Print("[WATCH]")
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("Put\tRevision: ", event.Kv.CreateRevision, event.Kv.ModRevision)
				res.ResType = PUT
				res.Key = key
				res.IsCreate = event.IsCreate()
				res.IsModify = event.IsModify()
				res.ValueBytes = event.Kv.Value
				data, _ := json.Marshal(res)
				publisher, _ := message.NewPublisher(config.DefaultQConfig())
				publisher.Publish("/testAddPod", data, "application/json")
				break
			case mvccpb.DELETE:
				res.ResType = DELETE
				fmt.Println("Delete\tRevision:", event.Kv.ModRevision)
				break
=======
	watch := func(c chan<- WatchRes) {
		//fmt.Println("watch again")
		watchRespChan := watcher.Watch(ctx, key)
		// 处理kv变化事件
		for watchResp := range watchRespChan {
			var res WatchRes
			for _, event := range watchResp.Events {
				fmt.Print("[WATCH-RESULT]")
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("Put\tRevision: ", event.Kv.CreateRevision, event.Kv.ModRevision)
					data, _ := json.Marshal("sewgwq")
					publisher, _ := message.NewPublisher(message.DefaultQConfig())
					publisher.Publish("/testwatch", data, "application/json")
				case mvccpb.DELETE:
					fmt.Println("Delete\tRevision:", event.Kv.ModRevision)
				}
>>>>>>> 22c9598c726809453582ec62946941ca15843c80
			}
		}
	}

	return cancel, watchResChan
}
