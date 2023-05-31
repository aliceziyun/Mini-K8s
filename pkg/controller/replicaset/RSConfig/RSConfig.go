package RSConfig

import "Mini-K8s/pkg/controller/replicaset/RSUpdate"

type RSConfig struct {
	updates chan RSUpdate.RSUpdate
}

func NewRSConfig() *RSConfig {
	updates := make(chan RSUpdate.RSUpdate)
	rsconfig := &RSConfig{updates: updates}
	return rsconfig
}

func (rc *RSConfig) GetUpdates() chan RSUpdate.RSUpdate {
	return rc.updates
}
