package kubeproxy

import (
	"Mini-K8s/pkg/iptable"
)

type KubeProxy struct{}

func NewKubeProxy() *KubeProxy {
	kubeProxy := &KubeProxy{}
	return kubeProxy
}

func (kubeProxy *KubeProxy) Run() {
	//先判断services链存不存在
	ipt, _ := iptable.New()
	exist, _ := ipt.isChainExist()

	if exist {
		return
	}

	//创建该链并做处理
	_ = ipt.NewChain()
	_ = ipt.Insert()
	_ = ipt.Insert()
}
