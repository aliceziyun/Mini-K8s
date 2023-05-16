package kubeproxy

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/iptable"
	"Mini-K8s/pkg/listener"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"time"
)

type KubeProxy struct {
	ls          *listener.Listener
	stopChannel <-chan struct{}
}

func NewKubeProxy() *KubeProxy {
	kubeProxy := &KubeProxy{}
	return kubeProxy
}

func (kubeProxy *KubeProxy) Run() {
	initService()
	watchService := func() {
		for {
			err := kubeProxy.ls.Watch(etcdstorage.EtcdServicePrefix, kubeProxy.serviceChangeHandler, kubeProxy.stopChannel)
			if err != nil {
				fmt.Println("[KubeProxy] watch error" + err.Error())
				time.Sleep(1 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
}

func initService() {

	//ipt, err := iptables.New()
	//if err != nil {
	//	fmt.Println("[chain] Boot error")
	//	fmt.Println(err)
	//}

	//fmt.Println("iptables -I FORWARD -i ens3 -j ACCEPT")
	//fmt.Println("iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE")
}

func testIpt() {

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

	fmt.Println("iptables -t nat -A OUTPUT -p tcp --dport 8000 -m statistic --mode random --probability 0.5 -j REDIRECT --to-ports 8080\niptables -t nat -A OUTPUT -p tcp --dport 8000 -j REDIRECT --to-ports 8081")
}

func (kubeProxy *KubeProxy) serviceChangeHandler(res etcdstorage.WatchRes) {
	if res.ResType == 0 {
		// delete service
	} else {
		// assume only exist CREAT
		service := &object.Service{}
		err := json.Unmarshal(res.ValueBytes, service)
		if err != nil {
			fmt.Println("[kubeProxy] Unmarshall fail" + err.Error())
			return
		}
		fmt.Println(service)

		ipt, err := iptables.New()
		if err != nil {
			fmt.Println("[chain] Boot error")
			fmt.Println(err)
		}

		dst1 := service.Spec.PodNameAndIps[0].Ip + ":" + service.Spec.Ports[0].TargetPort
		dst2 := service.Spec.PodNameAndIps[1].Ip + ":" + service.Spec.Ports[0].TargetPort

		// 将本机发往 10.10.0.2:8081 的包均匀地转发给 192.168.1.15:6666 和 192.168.1.4:8082
		fmt.Println("iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
			" -p tcp --dport " + service.Spec.Ports[0].Port +
			" -m statistic --mode random --probability 0.5" +
			" -j DNAT --to-destination " + dst1)
		fmt.Println("iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
			" -p tcp --dport " + service.Spec.Ports[0].Port +
			" -j DNAT --to-destination " + dst2)

		err = ipt.Append("nat", "OUTPUT",
			"--dst", service.Spec.ClusterIp, "-p", "tcp", "--dport", service.Spec.Ports[0].Port,
			"-m", "statistic", "--mode", "random", "--probability", "0.5",
			"-j", "DNAT", "--to-destination", dst1)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ipt.Append("nat", "OUTPUT",
			"--dst", service.Spec.ClusterIp, "-p", "tcp", "--dport", service.Spec.Ports[0].Port,
			"-j", "DNAT", "--to-destination", dst2)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func TestService(service object.Service) {
	dst1 := service.Spec.PodNameAndIps[0].Ip + ":" + service.Spec.Ports[0].TargetPort
	dst2 := service.Spec.PodNameAndIps[1].Ip + ":" + service.Spec.Ports[0].TargetPort

	// 将本机发往 10.10.0.2:8081 的包均匀地转发给 192.168.1.15:6666 和 192.168.1.4:8082
	fmt.Println("iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
		" -p tcp --dport " + service.Spec.Ports[0].Port +
		" -m statistic --mode random --probability 0.5" +
		" -j DNAT --to-destination " + dst1)
	fmt.Println("iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
		" -p tcp --dport " + service.Spec.Ports[0].Port +
		" -j DNAT --to-destination " + dst2)
}