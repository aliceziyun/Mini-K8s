package podConfig

import "Mini-K8s/pkg/kubelet/PodUpdate"

type PodConfig struct {
	// //podManager 与 dockerClient交互
	// //podManager *podManager.PodManager
	// podLock sync.RWMutex
	// //建立映射  source(pod源)--(map<pod id - *pod>)
	// pods map[string]map[string]*object.Pod
	updates chan PodUpdate.PodUpdate
}

// NewPodConfig TODO: complete new pod configuration
func NewPodConfig() *PodConfig {
	updates := make(chan PodUpdate.PodUpdate, 50)
	podConfig := &PodConfig{
		updates: updates,
	}
	return podConfig
}

func (pc PodConfig) GetUpdates() chan PodUpdate.PodUpdate {
	return pc.updates
}
