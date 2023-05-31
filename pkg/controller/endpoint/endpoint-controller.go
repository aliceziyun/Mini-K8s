package endpoint

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/third_party/queue"
	"context"
	"fmt"
)

type EndpointController struct {
	ls          *listwatcher.ListWatcher
	stopChannel chan struct{}
	queue       queue.ConcurrentQueue
}

func NewEndpointController(controllerContext controller_context.ControllerContext) *EndpointController {
	epc := &EndpointController{}
	epc.stopChannel = make(chan struct{})
	epc.ls = controllerContext.Ls
	return epc
}

func (epc *EndpointController) Run(ctx context.Context) {
	fmt.Println("[Endpoint Controller] start run ...")
	go epc.register()
	go epc.worker()
}

func (epc *EndpointController) register() {
	go func() {
		err := epc.ls.Watch(_const.SERVICE_CONFIG_PREFIX, epc.handleService, epc.stopChannel)
		if err != nil {
			fmt.Println("[Endpoint Controller] list watch RS handler init fail")
		}
	}()
}

func (epc *EndpointController) worker() {
	for {
		if !epc.queue.Empty() {
			key := epc.queue.Front()
			epc.queue.Dequeue()
			go func() {
				err := epc.syncService(key.(string))
				if err != nil {
					fmt.Println("[Endpoint Controller] worker error")
				}
			}()
		}
	}
}

func (epc *EndpointController) handleService(res etcdstorage.WatchRes) {

}

func (epc *EndpointController) syncService(key string) error {
	//获取service对象，当查询不到该service对象时，删除同名endpoints对象

	//根据service对象的Selector，查询与service对象匹配的pod列表

	//遍历service对象匹配的pod列表，找出合适的pod，计算endpoints的subsets信息

	//获取endpoints对象

	//当endpoints对象不存在时新建endpoints对象，当endpoints对象存在时更新endpoints对象
	return nil
}
