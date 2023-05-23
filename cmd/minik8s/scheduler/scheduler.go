package main

import (
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/scheduler"
	"fmt"
)

func main() {
	RunScheduler()
}

func RunScheduler() {
	fmt.Println("[Scheduler] scheduler start")
	scheduler.NewScheduler(listwatcher.DefaultConfig()).Run()
}
