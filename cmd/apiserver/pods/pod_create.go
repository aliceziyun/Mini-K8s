package pods

// import (
// 	"encoding/json"
// 	"fmt"
// 	clientv3 "go.etcd.io/etcd/client/v3"
// 	cliv3 "github.com/coreos/etcd/clientv3"
// 	"Mini-K8s/pkg/object"
// 	"Mini-K8s/pkg/etcdstorage"
// 	"strconv"
// )

// func CreatePod(cli *clientv3.Client, pod_ object.Pod) object.PodInstance {//PodInstance?
// 	podInstance := def.PodInstance{}
// 	podInstance.Pod = pod_

// 	// 先将新创建的pod写入到etcd当中
// 	podKey := "/pod/" + pod_.Metadata.Name
// 	podValue, err := json.Marshal(pod_)
// 	if err != nil {
// 		fmt.Printf("%v\n", err)
// 		panic(err)
// 	}
// 	etcd.Put(cli, podKey, string(podValue))

// 	//创建pod实例，写入etcd

// 	//podInstance.NodeID = nodeID
// 	podInstanceKey := def.GetKeyOfPodInstance(pod_.Metadata.Name)
// 	podInstance.ID = podInstanceKey
// 	podInstance.ContainerSpec = make([]def.ContainerStatus, len(pod_.Spec.Containers))

// 	podInstanceValue, err := json.Marshal(podInstance)
// 	if err != nil {
// 		fmt.Printf("%v\n", err)
// 		panic(err)
// 	}
// 	etcd.Put(cli, podInstanceKey, string(podInstanceValue))

// 	// //更新PodInstanceIDList
// 	// podInstanceIDList := make([]string, 0)
// 	// kvs := etcd.Get(cli, def.PodInstanceListID).Kvs
// 	// if len(kvs) != 0 {
// 	// 	podInstanceIDListValue := kvs[0].Value
// 	// 	err := json.Unmarshal(podInstanceIDListValue, &podInstanceIDList)
// 	// 	if err != nil {
// 	// 		fmt.Printf("%v\n", err)
// 	// 		panic(err)
// 	// 	}
// 	// }
// 	// podInstanceIDList = append(podInstanceIDList, podInstance.ID)
// 	// podInstanceIDValue, err := json.Marshal(podInstanceIDList)
// 	// if err != nil {
// 	// 	fmt.Printf("%v\n", err)
// 	// 	panic(err)
// 	// }
// 	// etcd.Put(cli, def.PodInstanceListID, string(podInstanceIDValue))

// 	return podInstance
// }
