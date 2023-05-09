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

	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}

	//fmt.Println("iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-ports 8000") // 外网转发

	fmt.Println("iptables -t nat -A OUTPUT -p tcp --dport 8000 -j REDIRECT --to-ports 8080") // 内网转发
	err = ipt.Append("nat", "OUTPUT", "-p", "tcp", "--dport", "8001", "-j", "REDIRECT", "--to-ports", "8080")
	if err != nil {
		fmt.Println(err)
		return
	}

}
