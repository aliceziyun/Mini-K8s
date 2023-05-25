package object

const (
	// kind
	POD        string = "POD"
	REPLICASET string = "REPLICASET"
	SERVICE    string = "SERVICE"
	HPA        string = "HPA"
)

const (
	// status
	RUNNING string = "RUNNING"
	STOP    string = "STOP"
	DELETED string = "DELETED"
)

type ObjMetadata struct {
	Name           string            `json:"name" yaml:"name"`
	Labels         map[string]string `json:"labels" yaml:"labels"`
	Uid            string            `json:"uid" yaml:"uid"`
	Namespace      string            `json:"namespace" yaml:"namespace"`
	OwnerReference []OwnerReference  `json:"ownerReferences" yaml:"ownerReferences"`
}

type OwnerReference struct {
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
	UID        string `json:"uid" yaml:"uid"`
	Controller bool   `json:"controller" yaml:"controller"` //指向controller的指针
}

// --------------------Container---------------------------
type Container struct {
	Name         string          `json:"name" yaml:"name"`
	Image        string          `json:"image" yaml:"image"`
	Ports        []ContainerPort `json:"ports" yaml:"ports"`
	Env          []ContainerEnv  `json:"env" yaml:"env"`
	Command      []string        `json:"command" yaml:"command"` // 容器的启动命令列表
	Args         []string        `json:"args" yaml:"args"`       // 容器的启动命令参数列表
	VolumeMounts []VolumeMount   `json:"volumeMounts" yaml:"volumeMounts"`
}

type VolumeMount struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
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
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
}

type ContainerPort struct {
	//added ?
	Name     string `json:"name" yaml:"name"`
	Port     string `json:"containerPort" yaml:"containerPort"`
	HostPort string `json:"hostPort" yaml:"hostPort"`
	//类型有三种 tcp, udp, all.      默认为tcp, all的话两种都开
	Protocol string `json:"protocol" yaml:"protocol"`
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

// ---------------------Pod-----------------------
type Pod struct {
	Name       string      `json:"name" yaml:"name"`
	ApiVersion int         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Metadata   ObjMetadata `json:"metadata" yaml:"metadata"`
	Spec       PodSpec     `json:"spec" yaml:"spec"`
	Status     PodStatus   `json:"status" yaml:"status"`
}

type PodSpec struct {
	Containers []Container `json:"containers" yaml:"containers"`
	Volumes    []Volume    `json:"volumes" yaml:"volumes"`
	NodeName   string      `json:"nodeName" yaml:"nodeName"`
}

type PodStatus struct {
	RunningContainers int32       `json:"running-containers" yaml:"running-containers"`
	Phase             string      `json:"phase" yaml:"phase"`
	IP                string      `json:"ip" yaml:"ip"`
	Conditions        []Condition `json:"conditions" yaml:"conditions"`
}

type PodNameAndIp struct {
	Name string `json:"name"`
	Ip   string `json:"ip"`
}

// ------------------------Node------------------------
type Node struct {
	MetaData ObjMetadata `json:"metadata" yaml:"metadata"`
	MasterIp string      `json:"masterIp" yaml:"masterIp"`
	Spec     NodeSpec    `json:"spec" yaml:"spec"`
	Status   string      `json:"status" yaml:"status"`
}

type NodeList struct {
	Items []Node `json:"items" yaml:"items"`
}

type NodeSpec struct {
	//浮动ip地址
	DynamicIp string `json:"physicalIp" yaml:"physicalIp""`
	//为该节点分配的pod网段
	NodeIpAndMask string `json:"nodeIpAndMask" yaml:"nodeIpAndMask"`
}

// ---------------------ReplicaSet----------------------
type ReplicaSet struct {
	ObjMetadata `json:"metadata" yaml:"metadata"`
	Spec        ReplicaSetSpec   `json:"spec" yaml:"spec"`
	Status      ReplicaSetStatus `json:"status" yaml:"status"`
}

type ReplicaSetSpec struct {
	Replicas int32 `json:"replicas" yaml:"replicas"`
	Pods     Pod   `json:"template" yaml:"template"`
}

type ReplicaSetStatus struct {
	ReplicaStatus int32  `json:"replicas" yaml:"replicas"` //是否符合对replica的期待
	Status        string `json:"status" yaml:"status"`
}

// --------------------AutoScaler---------------------------
type Autoscaler struct {
	Metadata ObjMetadata `json:"metadata" yaml:"metadata"`
	Spec     HPASpec     `json:"spec" yaml:"spec"`
}

type HPASpec struct {
	ScaleTargetRef HPARef   `json:"scaleTargetRef" yaml:"scaleTargetRef"`
	MinReplicas    int32    `json:"minReplicas" yaml:"minReplicas"`
	MaxReplicas    int32    `json:"maxReplicas" yaml:"maxReplicas"`
	ScaleInterval  int32    `json:"scaleInterval" yaml:"scaleInterval"`
	Metrics        []Metric `json:"metrics" yaml:"metrics"`
}

type HPARef struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
}

type Metric struct {
	Name   string `json:"name" yaml:"name"`
	Target int    `json:"target" yaml:"target"`
}

// --------------------Service---------------------------
type Service struct {
	Name       string        `json:"name" yaml:"name"`
	ApiVersion int           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string        `json:"kind" yaml:"kind"`
	Metadata   ObjMetadata   `json:"metadata" yaml:"metadata"`
	Spec       ServiceSpec   `json:"spec" yaml:"spec"`
	Status     ServiceStatus `json:"status" yaml:"status"`
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

type ServiceStatus struct {
	Phase          string            `json:"phase" yaml:"phase"`
	Pods2IpAndPort map[string]string `json:"pods2IpAndPort" yaml:"pods2IpAndPort"` //pod name到 podIp:port的映射
}
