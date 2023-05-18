package controller

import (
	"Mini-K8s/pkg/controller/deployment/config"
	"Mini-K8s/pkg/controller/replicaset/config"
)

type Config struct {
	*replicaset.ReplicaSetControllerOptions
	*deployment.DeploymentControllerOptions
}

type ControllerOptions struct {
	// TODO : add more controllers here
	ReplicaSetController *replicaset.ReplicaSetControllerOptions
	DeploymentController *deployment.DeploymentControllerOptions
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() *CompletedConfig {
	return &CompletedConfig{c}
}

func (option *ControllerOptions) Config() *Config {
	return &Config{
		ReplicaSetControllerOptions: option.ReplicaSetController,
		DeploymentControllerOptions: option.DeploymentController,
	}
}

func NewKubeControllerManagerOptions() *ControllerOptions {
	controllerManagerOptions := ControllerOptions{
		&replicaset.ReplicaSetControllerOptions{},
		&deployment.DeploymentControllerOptions{},
	}
	// TODO:这里记得设置点具体的值，虽然不确定这个到底有没有用
	return &controllerManagerOptions
}
