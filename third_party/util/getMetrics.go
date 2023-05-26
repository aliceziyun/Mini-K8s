package util

import (
	"fmt"
	"github.com/docker/docker/api/types"
)

//已经做了*100操作，换算成百分比

func GetCPUPercent(statsJson *types.StatsJSON) float64 {
	// CPU 利用率（%）=（容器 CPU 利用率的总变化/主机 CPU 利用率的变化）* 100
	// cpu_delta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
	// system_cpu_delta = cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage
	// CPU usage % = (cpu_delta / system_cpu_delta) * number_cpus * 100.0
	var preCPUUsage uint64
	var CPUUsage uint64
	preCPUUsage = 0
	CPUUsage = 0
	for _, core := range statsJson.CPUStats.CPUUsage.PercpuUsage {
		CPUUsage += core
	}
	for _, core := range statsJson.PreCPUStats.CPUUsage.PercpuUsage {
		preCPUUsage += core
	}
	systemUsage := statsJson.CPUStats.SystemUsage
	preSystemUsage := statsJson.PreCPUStats.SystemUsage

	deltaCPU := CPUUsage - preCPUUsage
	deltaSystem := systemUsage - preSystemUsage

	onlineCPU := statsJson.CPUStats.OnlineCPUs

	cpuPercent := (float64(deltaCPU) / float64(deltaSystem)) * float64(onlineCPU) * 100.0
	return cpuPercent
}

func GetMemPercent(statsJson *types.StatsJSON) float64 {
	// MemPercent = USAGE / LIMIT
	usage := statsJson.MemoryStats.Usage
	maxUsage := statsJson.MemoryStats.Limit
	percentage := float64(usage) / float64(maxUsage) * 100.0
	return percentage
}

func PrintMetricJson(byte_ []byte) {
	newStr := string(byte_)
	fmt.Println("containerStats.Body:")
	fmt.Println(newStr)
}
