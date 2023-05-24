package main

import (
	"Mini-K8s/pkg/shell"
	"fmt"
)

func main() {
	output, err := shell.ExecCmd("cat", "/etc/hosts")
	if err != nil {
		return
	}
	fmt.Println(output)
}
