package main

import (
	"Mini-K8s/pkg/mesh"
)

func main() {
	sidecar := mesh.Sidecar{
		PodIP: "10.10.72.2",
		Host:  "192.168.1.4",
	}
	//sidecar.RunForwardServer("outbound")
	sidecar.RunForwardServer("inbound")
}
