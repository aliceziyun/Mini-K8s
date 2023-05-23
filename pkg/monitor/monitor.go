package monitor

//var (
//	podMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
//		Name: "pod_metric",
//	}, []string{"resource", "pod", "uid"})
//)
//
//type Monitor struct {
//	dockerClient *client.Client
//}
//
//func (m *Monitor) Listener() {
//	http.Handle("/metrics", promhttp.Handler())
//	err := http.ListenAndServe(":2112", nil)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//}
//
//func (m *Monitor) GetDockerStat(ctx context.Context, pod *pod.Pod) {
//	if pod == nil {
//		return
//	}
//	containers := pod.GetContainers()
//	for _, container := range containers {
//		containerID := container.ContainerId
//		stats, err := m.dockerClient.ContainerStats(ctx, containerID, false)
//		if err != nil {
//			fmt.Printf("[Monitor] Get stats error:%v\n", err)
//			continue
//		}
//		raw, _ := ioutil.ReadAll(stats.Body)
//
//		statsJson := &types.StatsJSON{}
//		err = json.Unmarshal(raw, statsJson)
//		if err != nil {
//			fmt.Println("[Monitor]", err)
//			continue
//		}
//
//		cpuPercent := util.GetCPUPercent(statsJson)
//		memPercent := util.GetMemPercent(statsJson)
//
//		NewMetric(pod.GetName(), pod.GetUid(), memPercent, cpuPercent)
//	}
//}
//
//func NewMetric(podName string, podUID string, memPercent float64, cpuPercent float64) {
//	podMetric.WithLabelValues("memory", podName, podUID).Set(memPercent)
//	podMetric.WithLabelValues("cpu", podName, podUID).Set(cpuPercent)
//}
