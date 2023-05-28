package main

import "Mini-K8s/pkg/mesh"

func main() {
	sidecar := mesh.Sidecar{}
	sidecar.RunForwardServer("192.168.1.4:15001", "10.10.72.2", "-s")
}
