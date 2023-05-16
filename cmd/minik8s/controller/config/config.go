package config

import "Mini-K8s/pkg/controller/deployment/config"

type Config struct {
	*config.DeploymentControllerOptions
	//*controllers.AutoscalerControllerOptions
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() *CompletedConfig {
	// TODO : complete podConfig
	return &CompletedConfig{c}
}
