package autoscaler

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type AutoScaleController struct {
	ls           *listwatcher.ListWatcher
	stopChannel  <-chan struct{}
	queue        queue.ConcurrentQueue
	hashMap      *_map.ConcurrentMap
	metricClient client.MetricClient
}

func NewAutoScaleController(controllerContext controller_context.ControllerContext) *AutoScaleController {
	hash := _map.NewConcurrentMap()
	mClient := client.MetricClient{Base: "localhost:8080"}
	asc := &AutoScaleController{
		ls:           controllerContext.Ls,
		hashMap:      hash,
		metricClient: mClient,
	}
	return asc
}

func (asc *AutoScaleController) Run(ctx context.Context) {
	fmt.Println("[AutoScale Controller] start run ...")
	go asc.worker()

	<-asc.stopChannel
}

func (asc *AutoScaleController) worker() {
	//在运行processNextWorkItem后，每隔一秒执行一次worker
	//这里和原代码有一些出入，源代码是固定一秒开启一个worker协程，这里简化了
	for {
		for asc.processNextWorkItem() {
			//没有workItem需要处理时，退出循环
		}
		time.Sleep(time.Second * 1)
	}
}

// 遍历所有对象
func (asc *AutoScaleController) processNextWorkItem() bool {
	var m sync.Mutex
	if !asc.queue.Empty() {
		return false
	} else {
		key := asc.queue.Front()
		m.Lock()
		_, err := asc.reconcileKey(key.(string))
		if err != nil {
			fmt.Println(err)
		}
		m.Unlock()

		return true
	}
}

func (asc *AutoScaleController) reconcileKey(key string) (deleted bool, err error) {
	if hpa, ok := asc.hashMap.Get(key).(*object.Autoscaler); !ok {
		err := errors.New("[AutoScale Controller] HPA has been deleted")
		return true, err
	} else {
		return false, asc.reconcileAutoscaler(hpa, key)
	}
}

func (asc *AutoScaleController) reconcileAutoscaler(autoscaler *object.Autoscaler, key string) error {
	//根据autoScaler选择出应有的replicaSet资源
	targetName := autoscaler.Spec.ScaleTargetRef.Name
	rsList, err := asc.ls.List(_const.RS_CONFIG_PREFIX + "/" + targetName)
	if err != nil {
		return err
	}
	rs := &object.ReplicaSet{}
	err = json.Unmarshal(rsList[0].ValueBytes, &rs)
	if err != nil {
		return err
	}

	var desiredReplicas int32
	currentReplicas := rs.Status.ReplicaStatus
	minReplicas := autoscaler.Spec.MinReplicas
	maxReplicas := autoscaler.Spec.MaxReplicas

	//副本数为0，不启动自动扩缩容
	if rs.Spec.Replicas == 0 && minReplicas != 0 {
		//TODO: disabled
		desiredReplicas = 0
		return nil
	} else if currentReplicas > maxReplicas { //感觉有一种和replicaSet冲突的美
		desiredReplicas = maxReplicas
	} else if currentReplicas < minReplicas {
		desiredReplicas = minReplicas
	} else {
		metricDesiredReplicas, metricName, metricTimestamp, err := asc.computeReplicasForMetrics(autoscaler, rs, autoscaler.Spec.Metrics)
		if err != nil {
			return err
		}
	}
	return nil
}

// 根据实际情况计算到底需要多少个Replica
func (asc *AutoScaleController) computeReplicasForMetrics(autoscaler *object.Autoscaler, rs *object.ReplicaSet,
	metrics []object.Metric) (replicas int32, metric string, timestamp time.Time, err error) {
	for _, metric := range metrics {
		replicaCountProposal, metricNameProposal, timestampProposal, err := asc.computeReplicasForMetric(autoscaler, metric, rs)
	}
}

// 根据某项metric计算需要多少个replica
func (asc *AutoScaleController) computeReplicasForMetric(autoscaler *object.Autoscaler, metric object.Metric,
	rs *object.ReplicaSet) (replicaCountProposal int, metricNameProposal string, timestampProposal time.Time, err error) {
	
}
