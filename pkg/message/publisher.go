package message

import (
	"fmt"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type Publisher struct {
	conn          *amqp.Connection
	connUrl       string
	maxRetry      int
	retryInterval time.Duration
	normal        bool
	mtxNormal     sync.Mutex
}

func NewPublisher(config *QConfig) (*Publisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", config.User, config.Password, config.Host, config.Port, config.VHost)
	p := new(Publisher)
	var err error
	p.connUrl = url
	p.maxRetry = config.MaxRetry
	p.retryInterval = config.RetryInterval
	p.normal = false
	p.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	errCh := make(chan *amqp.Error)
	//go p.rerun(errCh)
	p.conn.NotifyClose(errCh)
	return p, nil
}

func (p *Publisher) Publish(exchangeName string, body []byte, contentType string) error {
	fmt.Println(exchangeName)
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

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

	err = ch.Publish(
		exchangeName,
		exchangeName,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		})
	if err != nil {
		return err
	}
	return nil
}
