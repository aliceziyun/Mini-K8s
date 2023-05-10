package main

import (
	"Mini-K8s/pkg/kubeproxy"
)

func main() {
	kubeProxy := kubeproxy.NewKubeProxy()
	kubeProxy.Run()
}
