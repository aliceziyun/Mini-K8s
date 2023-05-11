package controller

import (
	"Mini-K8s/cmd/minik8s/controller/config"
	"Mini-K8s/pkg/listwatcher"
)

type ControllerContext struct {
	Ls             *listwatcher.ListWatcher
	MasterIP       string
	HttpServerPort string
	PromServerPort string
	Config         *config.CompletedConfig
}
