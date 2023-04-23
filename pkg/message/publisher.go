package message

import (
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
