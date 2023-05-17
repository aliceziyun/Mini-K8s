package _const

// build
const (
	PODFILE string = "/home/lcz/go/src/Mini-K8s/build/Pod/testPod.yaml"
)

// REST resource
const (
	ETCD_POD_PREFIX     string = "/registry/pods/"
	ETCD_SERVICE_PREFIX string = "/registry/services/"

	POD_CONFIG_PREFIX string = "/registry/pod/default"

	RS_CONFIG_PREFIX string = "/registry/rsConfig/default"
	RS_PREFIX        string = "/registry/rs/default"
)

// api-server
const (
	BASE_URI string = "localhost"
)
