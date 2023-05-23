package _const

const (
	SHARED_DATA_DIR = "/home/lcz/SharedData"
)

// build
const (
	PODFILE string = "/home/lcz/go/src/Mini-K8s/build/Pod/testPod.yaml"
	RSFILE  string = "/home/lcz/go/src/Mini-K8s/build/ReplicaSet/testRS.yaml"
	JOBFILE string = "/home/lcz/go/src/Mini-K8s/build/Job/testJob.yaml"
)

// REST resource
const (
	PATH string = "/registry/:resource/:namespace"

	POD_RUNTIME_PREFIX string = "/registry/pod/default"
	POD_CONFIG         string = "/registry/pod_config/default/:resourceName"
	POD_CONFIG_PREFIX  string = "/registry/pod_config/default"

	RS_CONFIG        string = "/registry/rs_config/default/:resourceName"
	RS_CONFIG_PREFIX string = "/registry/rs_config/default"

	SERVICE_CONFIG_PREFIX string = "/registry/service_config/default"

	JOB_CONFIG        string = "/registry/job/default/"
	JOB_CONFIG_PREFIX string = "/registry/job/default/:resourceName"

	SHARED_DATA        string = "/registry/sharedData/default/:resourceName"
	SHARED_DATA_PREFIX string = "/registry/sharedData/default"
)

// api-server
const (
	BASE_URI string = "http://localhost:8080"
)
