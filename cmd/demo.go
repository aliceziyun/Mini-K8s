package main

import (
	"Mini-K8s/pkg/mesh"
)

func main() {
	mesh.RunSidecar()
	select {}
}
