package mesh

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/pkg/selector"
	"encoding/json"
	"fmt"
	"time"
)

type Endpoint struct {
	Ip     string
	Weight string
}

type Gateway struct {
	ls           *listwatcher.ListWatcher
	stopChannel  <-chan struct{}
	Name2Service map[string]object.Service
	Ip2Endpoint  map[string][]Endpoint
}

func RunGateway(lsConfig *listwatcher.Config) *Gateway {
	gateway := &Gateway{}
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	gateway.ls = ls

	gateway.Name2Service = make(map[string]object.Service)
	gateway.Ip2Endpoint = make(map[string][]Endpoint)

	watchService := func() {
		for {
			err := gateway.ls.Watch(_const.SERVICE_CONFIG_PREFIX, gateway.serviceChangeHandler, gateway.stopChannel)
			if err != nil {
				fmt.Println("[Gateway] watch error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	watchVService := func() {
		for {
			err := gateway.ls.Watch(_const.VSERVICE_CONFIG_PREFIX, gateway.virtualServiceChangeHandler, gateway.stopChannel)
			if err != nil {
				fmt.Println("[Gateway] watch error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
	go watchVService()

	return gateway
}

func (g *Gateway) serviceChangeHandler(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		fmt.Println("Delete " + res.Key)
	} else {
		service := &object.Service{}
		err := json.Unmarshal(res.ValueBytes, service)
		if err != nil {
			fmt.Println("[Gateway] Unmarshall fail" + err.Error())
			return
		}
		g.Name2Service[service.Name] = *service
	}
}

func (g *Gateway) virtualServiceChangeHandler(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		fmt.Println("Delete " + res.Key)
	} else {
		vs := &object.VService{}
		err := json.Unmarshal(res.ValueBytes, vs)
		if err != nil {
			fmt.Println("[Gateway] Unmarshall fail" + err.Error())
			return
		}
		fmt.Println(vs)

		v2w := make(map[int]string)
		for _, val := range vs.Spec.PodVersionAndWeights {
			v2w[val.ApiVersion] = val.Weight
		}

		service := object.Service{}
		service = g.Name2Service[vs.Spec.ServiceName]
		servRuntime := selector.NewService(&service, listwatcher.DefaultConfig())

		var endpoints []Endpoint
		for _, pod := range servRuntime.Pods {
			weight, ok := v2w[pod.ApiVersion]
			if ok {
				endpoints = append(endpoints, Endpoint{pod.Status.PodIP, weight})
			}
		}

		//podNum := len(endpoints)
		//probability := strconv.FormatFloat(1/float64(podNum), 'f', 2, 64)

		g.Ip2Endpoint[service.Spec.ClusterIp] = endpoints

	}
}

func (g *Gateway) transferDstIp(ipv4 string, port uint16) (string, uint16) {
	endpoints, ok := g.Ip2Endpoint[ipv4]
	if !ok {
		// there is no virtual service for this ip
		return ipv4, port
	}

	// todo
	fmt.Println(endpoints)
	return ipv4, port
}
