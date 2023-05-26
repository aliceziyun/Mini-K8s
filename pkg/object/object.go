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

type ServiceSpec struct {
	Type          string            `json:"type" yaml:"type"`           //service 的类型，有ClusterIp和 NodePort类型,默认为ClusterIp,暂时只支持ClusterIp
	ClusterIp     string            `json:"clusterIp" yaml:"clusterIp"` //虚拟服务Ip地址， 可以手工指定或者由系统进行分配
	Ports         []ServicePort     `json:"ports" yaml:"ports"`         //service需要暴露的端口列表
	Selector      map[string]string `json:"selector" yaml:"selector"`
	PodNameAndIps []PodNameAndIp    `json:"podNameAndIps"` //选取的podsIp
}
type ServicePort struct {
	Name       string `json:"name" yaml:"name"`
	Protocol   string `json:"protocol" yaml:"protocol"`     //端口协议, 支持TCP和UDP, 默认TCP
	Port       string `json:"port" yaml:"port"`             //服务监听的端口号
	TargetPort string `json:"targetPort" yaml:"targetPort"` //需要转发到后端Pod的端口号
	NodePort   string `json:"nodePort" yaml:"nodePort"`     //当service类型为NodePort时，指定映射到物理机的端口号
}
type PodNameAndIp struct {
	Name string `json:"name"`
	Ip   string `json:"ip"`
}
type ServiceStatus struct {
	Phase          string            `json:"phase" yaml:"phase"`
	Pods2IpAndPort map[string]string `json:"pods2IpAndPort" yaml:"pods2IpAndPort"` //pod name到 podIp:port的映射
}

type Service struct {
	Name       string        `json:"name" yaml:"name"`
	ApiVersion int           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string        `json:"kind" yaml:"kind"`
	Metadata   ObjMetadata   `json:"metadata" yaml:"metadata"`
	Spec       ServiceSpec   `json:"spec" yaml:"spec"`
	Status     ServiceStatus `json:"status" yaml:"status"`
}

type Container struct {
	Name    string          `json:"name" yaml:"name"`
	Image   string          `json:"image" yaml:"image"`
	Ports   []ContainerPort `json:"ports" yaml:"ports"`
	Env     []ContainerEnv  `json:"env" yaml:"env"`
	Command []string        `json:"command" yaml:"command"`
	Args    []string        `json:"args" yaml:"args"`
}
type Containers struct {
	Containers []Container `json:"containers" yaml:"containers"`
}

// ContainerMeta (added)
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

func (pod *Pod) GetContainerByName(name string) *Container {
	for _, container := range pod.Spec.Containers {
		if container.Name == name {
			return &container
		}
	}
	return nil
}
func (pod *Pod) Containers() []Container {
	return pod.Spec.Containers
}
