package main

import (
	"Mini-K8s/cmd/minik8s/kubectl/commands"
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strings"
)

var app *cli.App

func main() {
	app = initApp()
	fmt.Println("welcome to use mini-k8s!")
	for {
		fmt.Printf(">")
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.Trim(cmd, "\r\n")
		parseArg(cmd)
	}
}

func initApp() *cli.App {
	if app == nil {
		app = cli.NewApp()
		app.Name = "Mini-K8s"
		app.Usage = "上海交通大学云操作系统大作业，简易版K8s"
		app.Commands = []cli.Command{
			commands.NewApplyCommand(),
			commands.NewQuitCommand(),
			commands.NewGetPodCommand(),
			commands.NewDeleteCommand(),
		}
	}
	return app
}

func parseArg(cmd string) {
	_ = app.Run(strings.Split(cmd, " "))
}
