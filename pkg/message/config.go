package message

import "time"

type QConfig struct {
	User          string
	Password      string
	Host          string
	Port          string
	VHost         string
	MaxRetry      int
	RetryInterval time.Duration
}

func DefaultQConfig() *QConfig {
	config := QConfig{
		User:          "root",
		Password:      "Asd11111",
		Host:          "localhost",
		Port:          "5672",
		VHost:         "admin",
		MaxRetry:      10,
		RetryInterval: 5 * time.Second,
	}
	return &config
}
