package message

import (
	"fmt"

	"github.com/streadway/amqp"
)

type HandleFunc func(d amqp.Delivery)

type Subscriber struct {
	Conn     *amqp.Connection
	ConnUrl  string
	MaxRetry int
}

func NewSubscriber(config *QConfig) (*Subscriber, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", config.User, config.Password, config.Host, config.Port, config.VHost)
	var err error

	s := new(Subscriber)
	s.ConnUrl = url
	s.MaxRetry = config.MaxRetry

	s.Conn, err = amqp.Dial(url)
	if err != nil {
		fmt.Println("Fail to connect to server")
		return nil, err
	}

	errCh := make(chan *amqp.Error)

	//go s.rerun(errCh)

	s.Conn.NotifyClose(errCh)
	return s, nil
}

func (s *Subscriber) Subscribe(exchangeName string, handler HandleFunc, stopChannelCh <-chan struct{}) error {
	fmt.Println("begin subscribe with name", exchangeName)
	ch, err := s.Conn.Channel()
	if err != nil {
		return err
	}

	// 进行绑定
	err = ch.ExchangeDeclare(
		exchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		queue.Name,
		exchangeName,
		exchangeName,
		false,
		nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	stopConnectionCh := make(chan *amqp.Error)
	s.Conn.NotifyClose(stopConnectionCh)
	go stop(ch, stopChannelCh, stopConnectionCh)
	go consumeLoop(msgs, handler)

	return nil
}

func stop(amqpChannel *amqp.Channel, stopChannelCh <-chan struct{}, stopConnectionCh <-chan *amqp.Error) {
	select {
	case <-stopConnectionCh:
		fmt.Println("connection closed!\n")
		return
	case <-stopChannelCh:
		_ = amqpChannel.Close()
		return
	}
}

// 真·处理函数
func consumeLoop(deliveries <-chan amqp.Delivery, handler HandleFunc) {
	for d := range deliveries {
		handler(d)
	}
}

func (s *Subscriber) Unsubscribe(stopChannelCh chan<- struct{}) {
	close(stopChannelCh)
}
