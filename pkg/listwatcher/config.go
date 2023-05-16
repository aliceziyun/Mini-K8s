package listwatcher

import (
	"Mini-K8s/pkg/message/config"
)

// Config :list-watch机制的config
type Config struct {
	Host        string
	Port        int
	QueueConfig *config.QConfig
}

func DefaultConfig() *Config {
	return &Config{
		Host:        "localhost",
		Port:        8080,
		QueueConfig: config.DefaultQConfig(),
	}
}
