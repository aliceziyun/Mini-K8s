package object

import (
	"fmt"
	"strings"
)

type GPUJob struct {
	Metadata ObjMetadata `json:"metadata" yaml:"metadata"`
	Spec     JobSpec     `json:"spec" yaml:"spec"`
}

type JobSpec struct {
	SlurmConfig JobConfig   `json:"slurm" yaml:"slurm"`
	App         AppTemplate `json:"template" yaml:"template"`
}

type JobConfig struct {
	JobName         string `json:"jobName" yaml:"jobName"`
	Partition       string `json:"partition" yaml:"partition"`
	Output          string `json:"output" yaml:"output"`
	Error           string `json:"error" yaml:"error"`
	Nodes           int32  `json:"N" yaml:"N"`
	NTasksPerNode   int32  `json:"nTasksPerNode" yaml:"nTasksPerNode"`
	CpusPerTask     int32  `json:"cpusPerTask" yaml:"cpusPerTask"`
	GenericResource string `json:"gres" yaml:"gres"`
}

type AppTemplate struct {
	AppSpec AppSpec `json:"spec" yaml:"spec"`
}

type AppSpec struct {
	Container     Container `json:"containers" yaml:"containers"`
	Commands      []string  `json:"command" yaml:"command"`
	ZipPath       string    `json:"zipPath" yaml:"zipPath"`
	RestartPolicy string    `json:"restartPolicy" yaml:"restartPolicy"`
}

type JobAppFile struct {
	Key   string
	Slurm []byte
	App   []byte
}

func (job *GPUJob) NewSlurmScript() []byte {
	var script []string
	slurmConfig := job.Spec.SlurmConfig
	script = append(script, "#!/bin/bash")
	script = append(script, fmt.Sprintf("#SBATCH --job-name=%s", slurmConfig.JobName))
	script = append(script, fmt.Sprintf("#SBATCH --partition=%s", slurmConfig.Partition))
	script = append(script, fmt.Sprintf("#SBATCH --output=%s", slurmConfig.Output))
	script = append(script, fmt.Sprintf("#SBATCH --error=%s", slurmConfig.Error))
	script = append(script, fmt.Sprintf("#SBATCH --n=%d", slurmConfig.Nodes))
	script = append(script, fmt.Sprintf("#SBATCH --ntasks-per-node=%d", slurmConfig.NTasksPerNode))
	script = append(script, fmt.Sprintf("#SBATCH --cpus-per-task=%d", slurmConfig.CpusPerTask))
	script = append(script, fmt.Sprintf("#SBATCH --gres=%s", slurmConfig.GenericResource))

	for _, cmd := range job.Spec.App.AppSpec.Commands {
		script = append(script, cmd)
	}

	return []byte(strings.Join(script, "\n"))
}
