package endpoint

import (
	"Mini-K8s/cmd/minik8s/controller/controller"
	"Mini-K8s/pkg/listwatcher"
)

type EndpointController struct {
	ls          *listwatcher.ListWatcher
	stopChannel chan struct{}
}

func NewEndpointController(controllerContext controller.ControllerContext) *EndpointController {

}

func (epc *EndpointController) Run() {

}
