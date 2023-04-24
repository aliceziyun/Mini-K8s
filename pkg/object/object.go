package object

type ObjMetadata struct {
	Name      string            `json:"name" yaml:"name"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
	Uid       string            `json:"uid" yaml:"uid"`
	Namespace string            `json:"namespace" yaml:"namespace"`
}

type Pod struct {
	Name       string      `json:"name" yaml:"name"`
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
	Phase      string      `json:"phase"`
	Conditions []Condition `json:"conditions" yaml:"conditions"`
}

type PodNodeSelector struct {
}

type Container struct {
	Name    string          `json:"name" yaml:"name"`
	Image   string          `json:"image" yaml:"image"`
	Ports   []ContainerPort `json:"ports" yaml:"ports"`
	Env     []ContainerEnv  `json:"env" yaml:"env"`
	Command []string        `json:"command" yaml:"command"`
	Args    []string        `json:"args" yaml:"args"`
}

// ContainerMeta added
type ContainerMeta struct {
	OriginName  string
	RealName    string
	ContainerId string
}

type Volume struct {
}

type ContainerPort struct {
	//added ?
	Name          string `json:"name" yaml:"name"`
	ContainerPort string `json:"containerPort" yaml:"containerPort"`
	HostPort      string `json:"hostPort" yaml:"hostPort"`
	//类型有三种 tcp, udp, all.      默认为tcp, all的话两种都开
	Protocol string `json:"protocol" yaml:"protocol"`
	//?
}

type ContainerEnv struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type Condition struct {
	LastProbeTime      string `json:"lastProbeTime" yaml:"lastProbeTime"`
	LastTransitionTime string `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Status             string `json:"status" yaml:"status"`
	Type               string `json:"type" yaml:"type"`
}

type Scheduler struct {
}
