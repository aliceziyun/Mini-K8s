package listwatcher

import (
	_const "Mini-K8s/cmd/const"
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
		Host:        _const.MATSTER_INNER_IP,
		Port:        8080,
		QueueConfig: config.DefaultQConfig(),
	}
}
