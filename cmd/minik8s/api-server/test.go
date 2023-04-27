package main

import apiserver "Mini-K8s/cmd/minik8s/api-server/server"

func main() {
	RunApiServer()
}

func RunApiServer() {
	serverConfig := apiserver.DefaultServerConfig()
	server, _ := apiserver.NewServer(serverConfig)
	err := server.Run()
	if err != nil {
		return
	}
}
