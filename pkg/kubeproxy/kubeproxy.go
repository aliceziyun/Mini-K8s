package kubeproxy

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/iptable"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type KubeProxy struct {
	ServName2Serv map[string]object.Service
	ls            *listwatcher.ListWatcher
	stopChannel   <-chan struct{}
}

func NewKubeProxy(lsConfig *listwatcher.Config) *KubeProxy {
	kubeProxy := &KubeProxy{}
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	kubeProxy.ls = ls
	RunDNS(lsConfig)
	return kubeProxy
}

func (kubeProxy *KubeProxy) Run() {
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	exist, err := ipt.MyExist("FORWARD", "-i", "ens3", "-j", "ACCEPT")
	if exist == false {
		fmt.Println("iptables -I FORWARD -i ens3 -j ACCEPT")
		err = ipt.InsertWithoutTable("FORWARD", "-i", "ens3", "-j", "ACCEPT")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	exist, err = ipt.MyExist("POSTROUTING", "-t", "nat", "-o", "ens3", "-j", "MASQUERADE")
	if exist == false {
		fmt.Println("iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE")
		err = ipt.AppendNAT("POSTROUTING", "-o", "ens3", "-j", "MASQUERADE")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	watchService := func() {
		for {
			err := kubeProxy.ls.Watch(_const.SERVICE_CONFIG_PREFIX, kubeProxy.serviceChangeHandler, kubeProxy.stopChannel)
			if err != nil {
				fmt.Println("[KubeProxy] watch error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
}

func (kubeProxy *KubeProxy) serviceChangeHandler(res etcdstorage.WatchRes) {
	fmt.Println("kubeProxy handle watch")
	if res.ResType == etcdstorage.DELETE {
		kubeProxy.deleteService(res.Key)
	} else {
		service := &object.Service{}
		err := json.Unmarshal(res.ValueBytes, service)
		if err != nil {
			fmt.Println("[kubeProxy] Unmarshall fail" + err.Error())
			return
		}
		fmt.Println(service)
		kubeProxy.deleteService(service.Name)
		kubeProxy.ServName2Serv[service.Name] = *service

		ipt, err := iptables.New()
		if err != nil {
			fmt.Println(err)
			return
		}

		num := len(service.Spec.PodNameAndIps)
		for i := 0; i < num; i++ {
			dst := service.Spec.PodNameAndIps[i].Ip + ":" + service.Spec.Ports[0].TargetPort
			probability := strconv.FormatFloat(1/float64(num-i), 'f', 2, 64)
			if i+1 == num {
				probability = "1"
			}

			fmt.Println("iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
				" -p tcp --dport " + service.Spec.Ports[0].Port +
				" -m statistic --mode random --probability " + probability +
				" -j DNAT --to-destination " + dst)

			err = ipt.AppendNAT("OUTPUT", "--dst", service.Spec.ClusterIp, "-p", "tcp", "--dport", service.Spec.Ports[0].Port,
				"-m", "statistic", "--mode", "random", "--probability", probability,
				"-j", "DNAT", "--to-destination", dst)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func (kubeProxy *KubeProxy) deleteService(name string) {
	service, ok := kubeProxy.ServName2Serv[name]
	if !ok {
		fmt.Println("[KubeProxy] delete empty service")
		return
	}

	ipt, err := iptables.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	num := len(service.Spec.PodNameAndIps)
	for i := 0; i < num; i++ {
		dst := service.Spec.PodNameAndIps[i].Ip + ":" + service.Spec.Ports[0].TargetPort
		probability := strconv.FormatFloat(1/float64(num-i), 'f', 2, 64)
		if i+1 == num {
			probability = "1"
		}

		fmt.Println("delete iptables -t nat -A OUTPUT --dst " + service.Spec.ClusterIp +
			" -p tcp --dport " + service.Spec.Ports[0].Port +
			" -m statistic --mode random --probability " + probability +
			" -j DNAT --to-destination " + dst)

		err = ipt.DeleteIfExists("nat", "OUTPUT", "--dst", service.Spec.ClusterIp, "-p", "tcp", "--dport", service.Spec.Ports[0].Port,
			"-m", "statistic", "--mode", "random", "--probability", probability,
			"-j", "DNAT", "--to-destination", dst)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func TestIpt() {

	fmt.Println("iptables -t nat -nL --line")
	fmt.Println("iptables -t nat -D PREROUTING 3")

	// 将外网发来的包转发到本地另一个端口
	fmt.Println("iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-ports 8000")

	// 将内网发来的包转发到本地另一个端口
	fmt.Println("iptables -t nat -A OUTPUT -p tcp --dport 8000 -j REDIRECT --to-ports 8080")

	fmt.Println("iptables -I FORWARD -i ens3 -j ACCEPT")
	fmt.Println("iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE")

	// 将所有发往192.168.1.6:8081（本机）的包转发给192.168.1.4:6666
	fmt.Println("iptables -t nat -A PREROUTING --dst 192.168.1.6 -p tcp --dport 8081 -j DNAT --to-destination 192.168.1.4:6666")
	fmt.Println("iptables -t nat -A POSTROUTING --dst 192.168.1.4 -p tcp --dport 6666 -j SNAT --to-source 192.168.1.6")

	// 将所有发往10.0.0.1的包改为发往192.168.1.4
	fmt.Println("iptables -A OUTPUT -t nat -d 10.10.0.1 -j DNAT --to-destination 192.168.1.4")

	fmt.Println("iptables -t nat -A OUTPUT -p tcp --dport 8000 -m statistic --mode random --probability 0.5 -j REDIRECT --to-ports 8080\n" +
		"iptables -t nat -A OUTPUT -p tcp --dport 8000 -j REDIRECT --to-ports 8081")
}
