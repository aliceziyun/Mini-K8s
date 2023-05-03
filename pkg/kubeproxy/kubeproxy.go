package kubeproxy

import (
	"Mini-K8s/pkg/iptable"
	"fmt"
)

type KubeProxy struct{}

func NewKubeProxy() *KubeProxy {
	kubeProxy := &KubeProxy{}
	return kubeProxy
}

func (kubeProxy *KubeProxy) Run() {

	//ipt, _ := iptable.New()
	//exist, _ := ipt.isChainExist()
	//
	//if exist {
	//	return
	//}
	//
	//_ = ipt.NewChain()
	//_ = ipt.Insert()
	//_ = ipt.Insert()

	//先判断services链存不存在
	ipt, err := iptable.New()
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	exist, err2 := ipt.IsChainExist("nat", "SERVICE")
	if err2 != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	if exist {
		return
	}
	//创建该链并做处理
	err = ipt.NewChain("nat", "SERVICE")
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	err = ipt.Insert("nat", "OUTPUT", 1, "-j", "SERVICE", "-s", "0/0", "-d", "0/0", "-p", "all")
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	err = ipt.Insert("nat", "PREROUTING", 1, "-j", "SERVICE", "-s", "0/0", "-d", "0/0", "-p", "all")
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
}
