package etcdstorage

import (
	"context"
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

func (kvs *KVStore) Get(key string) ([]ListRes, error) {
	fmt.Println("wtf", key)
	kv := clientv3.NewKV(kvs.client)
	response, err := kv.Get(context.TODO(), key)
	if err != nil {
		return []ListRes{}, err
	}

	if len(response.Kvs) != 0 {
		fmt.Println("[etcd] get a new", key)
		listRes := ListRes{
			ResourceVersion: response.Kvs[0].ModRevision,
			CreateVersion:   response.Kvs[0].CreateRevision,
			Key:             string(response.Kvs[0].Key),
			ValueBytes:      response.Kvs[0].Value,
		}
		return []ListRes{listRes}, nil
	} else {
		return []ListRes{}, nil
	}
}

func (kvs *KVStore) GetPrefix(key string) ([]ListRes, error) {
	kv := clientv3.NewKV(kvs.client)
	response, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return []ListRes{}, err
	}
	var ret []ListRes
	for _, kv := range response.Kvs {
		res := ListRes{
			ResourceVersion: kv.ModRevision,
			CreateVersion:   kv.CreateRevision,
			Key:             string(kv.Key),
			ValueBytes:      kv.Value,
		}
		ret = append(ret, res)
	}
	return ret, nil
}

func (kvs *KVStore) Put(key string, val string) error {
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Put(context.TODO(), key, val)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("[etcd] put a new", key, val)
	return err
}

func (kvs *KVStore) Del(key string) error {
	fmt.Println("delete a new pod", key)
	kv := clientv3.NewKV(kvs.client)
	_, err := kv.Delete(context.TODO(), key)
	return err
}

// 现在这个Watch，Watch的是Prefix
func (kvs *KVStore) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	watchResChan := make(chan WatchRes)
	watcher := clientv3.NewWatcher(kvs.client)
	ctx, cancel := context.WithCancel(context.TODO())

	// 处理kv变化事件
	watch := func(c chan<- WatchRes) {
		watchRespChan := watcher.Watch(ctx, key, clientv3.WithPrefix())
		for watchResp := range watchRespChan {
			var res WatchRes
			for _, event := range watchResp.Events {
				fmt.Print("[WATCH]")
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("Put Revision: ", event.Kv.CreateRevision, event.Kv.ModRevision)
					res.ResType = PUT
					res.Key = key
					res.IsCreate = event.IsCreate()
					res.IsModify = event.IsModify()
					res.ValueBytes = event.Kv.Value
					break
				case mvccpb.DELETE:
					res.ResType = DELETE
					fmt.Println("Delete Revision:", event.Kv.ModRevision)
					break
				}
				c <- res
			}
		}
		close(c)
	}

	go watch(watchResChan)

	fmt.Printf("[etcd]  %s return \n", key)
	return cancel, watchResChan
}
