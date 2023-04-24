package listener

import "Mini-K8s/pkg/message"

type Config struct {
	Host        string
	Port        int
	QueueConfig *message.QConfig
}

func DefaultConfig() *Config {
	return &Config{
		Host:        "localhost",
		Port:        8080,
		QueueConfig: message.DefaultQConfig(),
	}
}