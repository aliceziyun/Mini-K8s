package autoscaler

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/client"
	"Mini-K8s/pkg/controller/replicaset"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
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

	//计算所需的replica数量
	if rs.Spec.Replicas == 0 && minReplicas != 0 { //副本数为0，不启动自动扩缩容
		//TODO: disabled
		desiredReplicas = 0
		return nil
	} else if currentReplicas > maxReplicas { //感觉有一种和replicaSet冲突的美
		desiredReplicas = maxReplicas
	} else if currentReplicas < minReplicas {
		desiredReplicas = minReplicas
	} else {
		metricDesiredReplicas, metricName, err := asc.computeReplicasForMetrics(rs, autoscaler.Spec.Metrics)
		if err != nil {
			return err
		}
		if metricDesiredReplicas > maxReplicas {
			desiredReplicas = maxReplicas
		} else {
			desiredReplicas = metricDesiredReplicas
		}
		fmt.Printf("[AutoScale Controller] choose %s as the metric with %d replicas \n", metricName, desiredReplicas)
	}

	//根据计算出的replica数量缩容扩容
	err = asc.scaleReplica(desiredReplicas, rs)
	if err != nil {
		return err
	}

	return nil
}

// 根据实际情况计算到底需要多少个Replica
func (asc *AutoScaleController) computeReplicasForMetrics(rs *object.ReplicaSet,
	metrics []object.Metric) (replicas int32, metric string, err error) {
	var maxValue int32 = 0
	var metricName string
	for _, metric := range metrics {
		replicaCount, err := asc.computeReplicasForMetric(metric, rs)
		if err != nil {
			fmt.Printf("[AutoScale Controller] count metric %s fail \n", metric.Name)
		}
		//取最大值
		if replicaCount > maxValue {
			maxValue = replicaCount
			metricName = metric.Name
		}
	}
	return maxValue, metricName, err
}

// 根据某项metric计算需要多少个replica
func (asc *AutoScaleController) computeReplicasForMetric(metric object.Metric, rs *object.ReplicaSet) (replicaCount int32, err error) {
	//获取RS的全部pod
	pods, err := replicaset.GetAllPods(asc.ls, rs.Name, rs.Uid)
	if err != nil {
		return -1, err
	}

	//获取Pod的全部资源列表
	podResourceStatusList, err := asc.getPodResourceStatus(metric.Name, pods)
	if err != nil {
		return -1, err
	}
	for _, status := range podResourceStatusList {
		fmt.Printf("[AutoScale Controller] metric %s status is %s \n", metric.Name, status)
	}

	//计算需要的replica数量
	count, err := asc.computeReplicasCount(metric, podResourceStatusList)
	if err != nil {
		return -1, err
	}
	return count, err
}

func (asc *AutoScaleController) computeReplicasCount(metric object.Metric,
	resourceList []resourceStatus) (replicaCount int32, err error) {
	var allResource float64 = 0
	target := metric.Target
	for _, resource := range resourceList {
		allResource += resource.res * 100
	}
	if target == 0 {
		err := errors.New("[AutoScale Controller] target is 0")
		return -1, err
	}
	count := allResource / float64(target)
	return int32(math.Ceil(count)), nil
}

// 调整replicaSet中pod的数量
func (asc *AutoScaleController) scaleReplica(desireCount int32, rs *object.ReplicaSet) error {
	rs.Spec.Replicas = desireCount
	url := _const.BASE_URI + _const.RS_CONFIG_PREFIX + rs.Name
	payload, err := json.Marshal(rs)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(payload)
	request, err := http.NewRequest("PUT", url, reader)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("StatusCode not 200")
	}
	return nil
}
