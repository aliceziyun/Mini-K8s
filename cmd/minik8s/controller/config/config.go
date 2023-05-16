package controller

import (
	"Mini-K8s/pkg/controller/deployment/config"
	"Mini-K8s/pkg/controller/replicaset/config"
)

type Config struct {
	*replicaset.ReplicaSetControllerOptions
	*deployment.DeploymentControllerOptions
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() *CompletedConfig {
	return &CompletedConfig{c}
}
