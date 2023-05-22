package listwatcher

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/message"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"net/http"
)

type WatchHandler func(res etcdstorage.WatchRes)

type ListWatcher struct {
	Subscriber *message.Subscriber //指向subscriber的指针
	RootURL    string
}

// NewListWatcher :创建List-Watcher和与其绑定的subscriber
func NewListWatcher(c *Config) (*ListWatcher, error) {
	s, err := message.NewSubscriber(c.QueueConfig)
	if err != nil {
		return nil, err
	}
	ls := &ListWatcher{
		Subscriber: s,
		RootURL:    fmt.Sprintf("http://%s:%d", c.Host, c.Port),
	}
	return ls, nil
}

// List : 向API-Server发送一个http短链接请求，罗列所有目标资源的对象。
func (ls *ListWatcher) List(key string) ([]etcdstorage.ListRes, error) {
	fmt.Printf("[list watcher] list %s \n", key)
	resourceURL := ls.RootURL + key
	request, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}
	//向api-server发送请求
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var resList []etcdstorage.ListRes
	err = json.Unmarshal(data, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}

// Watch : 与某url长链接，监听某url绑定的操作，当对方有回复时，便调用watchHandler中的函数
func (l *ListWatcher) Watch(key string, handler WatchHandler, stopChannel <-chan struct{}) error {
	fmt.Printf("[list watcher] start watch %s \n", key)
	resourceURL := l.RootURL + key
	request, err := http.NewRequest("POST", resourceURL, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[list watcher] StatusCode not 200")
	}

	// 收到server的回复，开始监听
	stop := make(chan struct{})
	amqpHandler := func(d amqp.Delivery) {
		var res etcdstorage.WatchRes
		err := json.Unmarshal(d.Body, &res)
		if err != nil {
			fmt.Println("marshal error")
			return
		}
		handler(res)
	}

	err = l.Subscriber.Subscribe(key, amqpHandler, stop)
	if err != nil {
		return err
	}

	defer func() {
		l.Subscriber.Unsubscribe(stop)
	}()

	<-stopChannel
	return nil
}
