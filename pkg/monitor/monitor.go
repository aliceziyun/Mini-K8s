package monitor

import (
	"Mini-K8s/pkg/kubelet/pod"
	"Mini-K8s/third_party/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http"
)

var (
	podMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pod_metric",
	}, []string{"resource", "pod", "uid"})
)

type Monitor struct {
	dockerClient *client.Client
}

func NewMonitor() *Monitor {
	fmt.Printf("[Monitor] init... \n")
	c, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Println("[Monitor] init fail")
		return nil
	}

	return &Monitor{
		dockerClient: c,
	}
}

func (m *Monitor) Listener() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (m *Monitor) GetDockerStat(ctx context.Context, pod *pod.Pod) {
	if pod == nil {
		return
	}
	containers := pod.GetContainers()
	for _, container := range containers {
		containerID := container.ContainerId
		if containerID == "" {
			continue //还没填充，别急
		}
		fmt.Println("[Monitor] containerID,", containerID)
		stats, err := m.dockerClient.ContainerStats(ctx, containerID, false)
		if err != nil {
			fmt.Printf("[Monitor] Get stats error:%v\n", err)
			continue
		}
		raw, _ := ioutil.ReadAll(stats.Body)

		//util.PrintMetricJson(raw)

		statsJson := &types.StatsJSON{}
		err = json.Unmarshal(raw, statsJson)
		if err != nil {
			fmt.Println("[Monitor]", err)
			continue
		}

		cpuPercent := util.GetCPUPercent(statsJson)
		memPercent := util.GetMemPercent(statsJson)

		fmt.Println(cpuPercent, memPercent)

		NewMetric(pod.GetName(), pod.GetUid(), memPercent, cpuPercent)
	}
}

func NewMetric(podName string, podUID string, memPercent float64, cpuPercent float64) {
	podMetric.WithLabelValues("memory", podName, podUID).Set(memPercent)
	podMetric.WithLabelValues("cpu", podName, podUID).Set(cpuPercent)
}
