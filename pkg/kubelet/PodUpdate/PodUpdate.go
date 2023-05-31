package PodUpdate

import "Mini-K8s/pkg/object"

type PodUpdate struct {
	Pods   []*object.Pod
	Op     int64
	Source string
}
