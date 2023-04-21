package object

type ObjMetadata struct {
	Name      string            `json:"name" yaml:"name"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
	Uid       string            `json:"uid" yaml:"uid"`
	Namespace string            `json:"namespace" yaml:"namespace"`
}

type Pod struct {
	ApiVersion int         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Metadata   ObjMetadata `json:"metadata" yaml:"metadata"`
	Spec       PodSpec     `json:"spec" yaml:"spec"`
	Status     PodStatus   `json:"status" yaml:"status"`
}

type PodSpec struct {
	Containers   []Container     `json:"containers" yaml:"containers"`
	Volumes      []Volume        `json:"volumes" yaml:"volumes"`
	NodeSelector PodNodeSelector `json:"nodeSelector" yaml:"nodeSelector"`
}

type PodStatus struct {
	Conditions []Condition `json:"conditions" yaml:"conditions"`
}

type PodNodeSelector struct {
}

type Container struct {
	Name    string        `json:"name" yaml:"name"`
	Image   string        `json:"image" yaml:"image"`
	Ports   ContainerPort `json:"ports" yaml:"ports"`
	Env     ContainerEnv  `json:"env" yaml:"env"`
	Command string        `json:"command" yaml:"command"`
	Args    string        `json:"args" yaml:"args"`
}

type Volume struct {
}

type ContainerPort struct {
}

type ContainerEnv struct {
}

type Condition struct {
	LastProbeTime      string `json:"lastProbeTime" yaml:"lastProbeTime"`
	LastTransitionTime string `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Status             string `json:"status" yaml:"status"`
	Type               string `json:"type" yaml:"type"`
}
