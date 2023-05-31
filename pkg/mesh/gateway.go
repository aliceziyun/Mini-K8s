package mesh

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	iptables "Mini-K8s/pkg/iptable"
	"Mini-K8s/pkg/kubeproxy"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"time"
)

type Gateway struct {
	VServName2VServ map[string]object.VService
	ls              *listwatcher.ListWatcher
	stopChannel     <-chan struct{}
}

func NewGateway(lsConfig *listwatcher.Config) *Gateway {
	gateway := &Gateway{}
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	gateway.ls = ls
	watchService := func() {
		for {
			err := gateway.ls.Watch(_const.VSERVICE_CONFIG_PREFIX, gateway.vServiceChangeHandler, gateway.stopChannel)
			if err != nil {
				fmt.Println("[Gateway] watch error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
	return gateway
}

func (g *Gateway) vServiceChangeHandler(res etcdstorage.WatchRes) {
	fmt.Println("gateway handle watch")
	if res.ResType == etcdstorage.DELETE {
		vServName := res.Key
		vs := g.VServName2VServ[vServName]
		fmt.Println(vs)
	} else {
		vs := &object.VService{}
		err := json.Unmarshal(res.ValueBytes, vs)
		if err != nil {
			fmt.Println("[kubeProxy] Unmarshall fail" + err.Error())
			return
		}
		fmt.Println(vs)
		g.VServName2VServ[vs.Name] = *vs

		// find service

		// delete service
		p := kubeproxy.NewKubeProxy(listwatcher.DefaultConfig())
		p.DeleteService(vs.Spec.ServiceName)

		// add new rule
		ipt, err := iptables.New()
		if err != nil {
			fmt.Println(err)
			return
		}
		serviceIp := "10.10.0.8"
		servicePort := "6666"

		num := len(vs.Spec.PodIpAndWeights)
		for i := 0; i < num; i++ {
			ipt.AppendNAT("OUTPUT", "--dst", serviceIp, "-p", "tcp", "--dport", servicePort,
				"-m", "statistic", "--mode", "random", "--probability", vs.Spec.PodIpAndWeights[i].Weight,
				"-j", "DNAT", "--to-destination", vs.Spec.PodIpAndWeights[i].Ip)
		}
	}
}
