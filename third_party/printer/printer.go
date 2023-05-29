package printer

import (
	"Mini-K8s/pkg/object"
	"fmt"
)

func PrintPods(pods []object.UserPod) {
	fmt.Printf("NAME\tREADY\tSTATUS\tIP\n")

	for _, pod := range pods {
		fmt.Printf("%s\t%s\t%s\t%s\n", pod.Name, pod.Ready, pod.Status, pod.IP)
	}

	return
}

func PrintRS(rs []object.UserRS) {
	fmt.Printf("NAME\tCURRENT\tREADY\n")

	for _, each := range rs {
		fmt.Printf("%s\t%d\t%d\n", each.Name, each.Current, each.Ready)
	}

	return
}

func PrintNode(node []object.UserNode) {
	fmt.Printf("NAME\tDYNAMICIP\tCTIME\tROLE\n")

	for _, each := range node {
		fmt.Printf("%s\t%s\t%s\t%s\n", each.Name, each.DynamicIP, each.Ctime, each.Role)
	}
}
