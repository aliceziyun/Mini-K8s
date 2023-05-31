package controller_context

import (
	"Mini-K8s/cmd/minik8s/controller/config"
	"Mini-K8s/pkg/listwatcher"
)

type ControllerContext struct {
	Ls     *listwatcher.ListWatcher
	Config *controller.CompletedConfig
}
