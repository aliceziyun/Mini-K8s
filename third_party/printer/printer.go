package printer

import (
	"Mini-K8s/pkg/object"
	"fmt"
)

func PrintPods(pods []object.UserPod) {
	fmt.Printf("NAME\tREADY\tSTATUS\tIP\n")

	for _, pod := range pods {
		fmt.Printf("%s\t%s\t%s\t%s\t\n", pod.Name, pod.Ready, pod.Status, pod.IP)
	}

	return
}
