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

	fmt.Println("iptables -t nat -nL --line")
	fmt.Println("iptables -t nat -D PREROUTING 3")

	// 将外网发来的包转发到本地另一个端口
	fmt.Println("iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-ports 8000")

	// 将内网发来的包转发到本地另一个端口
	fmt.Println("iptables -t nat -A OUTPUT -p tcp --dport 8000 -j REDIRECT --to-ports 8080")
	err = ipt.Append("nat", "OUTPUT", "-p", "tcp", "--dport", "8001", "-j", "REDIRECT", "--to-ports", "8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("iptables -I FORWARD -i ens3 -j ACCEPT")
	fmt.Println("iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE")

	// 将所有发往192.168.1.6:8081（本机）的包转发给192.168.1.4:6666
	fmt.Println("iptables -t nat -A PREROUTING --dst 192.168.1.6 -p tcp --dport 8081 -j DNAT --to-destination 192.168.1.4:6666")
	fmt.Println("iptables -t nat -A POSTROUTING --dst 192.168.1.4 -p tcp --dport 6666 -j SNAT --to-source 192.168.1.6")
	err = ipt.Append("nat", "PREROUTING", "--dst", "192.168.1.6", "-p", "tcp", "--dport", "8081", "-j", "DNAT", "--to-destination", "192.168.1.4:6666")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ipt.Append("nat", "POSTROUTING", "--dst", "192.168.1.4", "-p", "tcp", "--dport", "6666", "-j", "SNAT", "--to-source", "192.168.1.6")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将所有发往10.0.0.1的包改为发往192.168.1.4
	fmt.Println("iptables -A OUTPUT -t nat -d 10.10.0.1 -j DNAT --to-destination 192.168.1.4")
	err = ipt.Append("nat", "OUTPUT", "-d", "10.10.0.1", "-j", "DNAT", "--to-destination", "192.168.1.4")
	if err != nil {
		fmt.Println(err)
		return
	}

}
