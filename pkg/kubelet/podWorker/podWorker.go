package podWorker

import (
	"Mini-K8s/pkg/kubelet/message"
)

// 用于Pod 与docker client之间的交互
type PodWorker struct {
}

func (podWorker *PodWorker) SyncLoop(commands <-chan message.PodCommand, responses chan<- message.PodResponse) {
	// for {
	// 	select {
	// 	case command, ok := <-commands:
	// 		if !ok {
	// 			return
	// 		}
	// 		res := dockerClient.HandleCommand(command.ContainerCommand)
	// 		result := message.PodResponse{
	// 			ContainerResponse: res,
	// 			PodResponseType:   command.PodCommandType,
	// 		}
	// 		responses <- result
	// 	}
	// }
}

// syncPod
