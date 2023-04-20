package etcdstorage

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type KVStore struct {
	client *clientv3.Client
}

// ref https://blog.csdn.net/wohu1104/article/details/108552649

func InitKVStore(endpoints []string, timeout time.Duration) (*KVStore, error) {
	fmt.Print("\n")
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	// establish a client
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	return &KVStore{client: cli}, nil
}

func (kvs *KVStore) Get(key string) error {
	kv := clientv3.NewKV(kvs.client)
	response, err := kv.Get(context.TODO(), key)
	if err != nil {
		return err
	}

	if len(response.Kvs) != 0 {
		fmt.Print("-> Get result:\n")
		fmt.Printf("\tkey: %s, value: %s\n", response.Kvs[0].Key, response.Kvs[0].Value)
	} else {
		fmt.Println("-> Get result: Empty")
	}
	fmt.Print("\n")
	return nil
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

func EtcdWatchTest() {
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"}, // 集群列表
		DialTimeout: 5 * time.Second,
	}

	// 建立一个客户端
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 获得kv API子集
	kv := clientv3.NewKV(client)

	// 模拟etcd中KV的变化
	go func() {
		for {
			kv.Put(context.TODO(), "/demo/A/B1", "i am B1")

			kv.Delete(context.TODO(), "/demo/A/B1")

			time.Sleep(1 * time.Second)
		}
	}()

	// 先GET到当前的值，并监听后续变化
	getResp, err := kv.Get(context.TODO(), "/demo/A/B1")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 当前etcd集群事务ID, 单调递增的
	watchStartRevision := getResp.Header.Revision + 1

	// 创建一个watcher
	watcher := clientv3.NewWatcher(client)

	// 启动监听
	fmt.Println("从该版本向后监听:", watchStartRevision)

	// 创建一个 5s 后取消的上下文
	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFunc()
	})

	// 该监听动作在 5s 后取消
	watchRespChan := watcher.Watch(ctx, "/demo/A/B1", clientv3.WithRev(watchStartRevision))

	// 处理kv变化事件
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改为:", string(event.Kv.Value), "Revision:",
					event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了", "Revision:", event.Kv.ModRevision)
			}
		}
	}

}

func EtcdTest() {
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"}, // 集群列表
		DialTimeout: 5 * time.Second,
	}

	// establish a client
	client, err := clientv3.New(config)
	if err != nil {
		return
	}

	kv := clientv3.NewKV(client)

	// put
	_, err = kv.Put(context.TODO(), "/etcd/1", "value1", clientv3.WithPrevKV())
	if err != nil {
		return
	}
	_, err = kv.Put(context.TODO(), "/etcd/2", "value2", clientv3.WithPrevKV())
	if err != nil {
		return
	}
	_, err = kv.Put(context.TODO(), "/etcd/3", "value3", clientv3.WithPrevKV())
	if err != nil {
		return
	}

	// read the key with the prefix "/demo/A"
	// clientv3.WithPrefix() , clientv3.WithCountOnly() can have more than one and just separate them with commas
	getResp, err := kv.Get(context.TODO(), "/etcd/", clientv3.WithPrefix() /*,clientv3.WithCountOnly()*/)
	if err != nil {
		return
	}

	fmt.Println(getResp.Kvs, getResp.Count)
	for _, resp := range getResp.Kvs {
		fmt.Printf("key: %s, value: %s\n", string(resp.Key), string(resp.Value))
	}

	// WithFromKey will delete the "/etcd/3" in this case
	_, err = kv.Delete(context.TODO(), "/etcd/2", clientv3.WithPrevKV() /*, clientv3.WithFromKey()*/)
	if err != nil {
		fmt.Println(err)
	}

	getResp1, err := kv.Get(context.TODO(), "/etcd/", clientv3.WithPrefix() /*,clientv3.WithCountOnly()*/)
	if err != nil {
		return
	}

	for _, resp := range getResp1.Kvs {
		fmt.Printf("key: %s, value: %s\n", string(resp.Key), string(resp.Value))
	}
}
