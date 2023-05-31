package main

import (
	fass_server "Mini-K8s/pkg/fass-server"
	"Mini-K8s/pkg/listwatcher"
)

func main() {
	RunFassServer()
}

func RunFassServer() {
	server := fass_server.NewServer(listwatcher.DefaultConfig())
	server.Run()
}
