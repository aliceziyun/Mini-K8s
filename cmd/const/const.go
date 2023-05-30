package _const

var NODE_NAME string

const (
	WORK_DIR        string = "./build"
	SHARED_DATA_DIR string = "/home/lcz/SharedData"
)

const (
	BASIC_IP_AND_MASK string = "172.16.0.0/16"
	MASTER_IP         string = "10.119.11.91"
	MATSTER_INNER_IP  string = "192.168.1.6"
)

// REST resource
const (
	PATH string = "/registry/:resource/:namespace/:resourceName"

	PATH_PREFIX string = "/registry/:resource/:namespace"

	POD_CONFIG         string = "/registry/pod_config/default/:resourceName"
	POD_CONFIG_PREFIX  string = "/registry/pod_config/default"
	POD_RUNTIME_PREFIX string = "/registry/pod/default"
	POD_META_PREFIX    string = "/registry/pod_meta/default"

	RS_CONFIG        string = "/registry/rs_config/default/:resourceName"
	RS_CONFIG_PREFIX string = "/registry/rs_config/default"

	SERVICE_CONFIG        string = "/registry/service_config/default/:resourceName"
	SERVICE_CONFIG_PREFIX string = "/registry/service_config/default"

	HPA_CONFIG_PREFIX string = "/registry/hpa_config/default"

	JOB_CONFIG        string = "/registry/job/default/:resourceName"
	JOB_CONFIG_PREFIX string = "/registry/job/default"

	NODE_CONFIG        string = "/registry/node/default/:resourceName"
	NODE_CONFIG_PREFIX string = "/registry/node/default"

	SHARED_DATA        string = "/registry/sharedData/default/:resourceName"
	SHARED_DATA_PREFIX string = "/registry/sharedData/default"

	DNS_CONFIG        string = "/registry/dns_config/default/:resourceName"
	DNS_CONFIG_PREFIX string = "/registry/dns_config/default"
)

// api-server
const (
	BASE_URI         string = "http://192.168.1.6:8080"
	BASE_MONITOR_URI string = "http://192.168.1.6:2112/metrics"
)
