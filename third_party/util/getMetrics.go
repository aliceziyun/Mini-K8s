package util

import "github.com/docker/docker/api/types"

func GetCPUPercent(statsJson *types.StatsJSON) float64 {
	// cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
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
