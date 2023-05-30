package printer

import (
	"Mini-K8s/pkg/object"
	"fmt"
)

func PrintPods(pods []object.UserPod) {
	fmt.Printf("NAME\tREADY\tSTATUS\tIP\tOWNER\tCTIME\n")

	for _, pod := range pods {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", pod.Name, pod.Ready, pod.Status, pod.IP, pod.Owner, pod.Ctime)
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
	fmt.Printf("NAME\tDYNAMICIP\tROLE\tCTIME\n")

	for _, each := range node {
		fmt.Printf("%s\t%s\t%s\t%s\n", each.Name, each.DynamicIP, each.Role, each.Ctime)
	}
}

func PrintSrv(srv []object.UserService) {
	for _, each := range srv {
		fmt.Println("Name: ", each.Name)
		fmt.Println("NameSpcace: ", each.NameSpace)
		fmt.Println("Selector: ", each.Selector)
		fmt.Println("IP Families: ", each.IPFamily)
		fmt.Println("IP: ", each.IP)
		fmt.Println("Port: ", each.Port)
		fmt.Println("Endpoint: ", each.EndPoint)
		fmt.Println("--------------------------------")
	}
}
