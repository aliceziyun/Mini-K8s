package main

import (
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/scheduler"
	"context"
	"fmt"
)

func main() {
	RunScheduler()
}

func RunScheduler() {
	fmt.Println("scheduler start")

	clientConfig := client.Config{Host: "127.0.0.1:8080"}
	sched := scheduler.NewScheduler(listwatcher.DefaultConfig(), clientConfig, "test")
	sched.Run(context.TODO())
	select {}
}
