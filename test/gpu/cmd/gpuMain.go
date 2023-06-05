package main

import (
	"Mini-K8s/test/gpu/pkg/gpu"
)

var (
	args = gpu.JobArgs{}
)

func main() {
	// 解析命令行参数
	// flag.StringVar(&args.JobName, "job-name", "gpu-job", "gpu job name")
	// flag.StringVar(&args.Output, "output", "output", "output filename")
	// flag.StringVar(&args.Error, "error", "error", "err filename")
	// flag.StringVar(&args.WorkDir, "workdir", "", "work directory")
	// flag.IntVar(&args.NumProcess, "process", 1, "number of processes(cpus)")
	// flag.IntVar(&args.NumTasksPerNode, "ntasks-per-node", 1, "number of tasks per node")
	// flag.IntVar(&args.CpusPerTask, "cpus-per-task", 1, "number of cpus per task")
	// flag.StringVar(&args.GpuResources, "gres", "gpu:1", "gpu resources")
	// flag.StringVar(&args.CompileScripts, "compile", "", "compile scripts")
	// flag.StringVar(&args.RunScripts, "run", "", "run scripts")
	// flag.StringVar(&args.Username, "username", "", "username")
	// flag.StringVar(&args.Password, "password", "", "password")
	// flag.Parse()

	// fmt.Printf("Read args: jobName = %v, output = %v, error = %v, n = %v\n", args.JobName, args.Output, args.Error, args.NumProcess)
	// fmt.Printf("Read args: numTasksPerNode = %v, cpusPerTask = %v, gpuResources = %v\n", args.NumTasksPerNode, args.CpusPerTask, args.GpuResources)
	// fmt.Printf("Read args: compileScripts = %s\n", args.CompileScripts)
	// fmt.Printf("Read args: runScripts = %s\n", args.RunScripts)
	// fmt.Printf("Read args: username = %s, password = %s\n", args.Username, args.Password)
	// fmt.Printf("Read args: workdir = %s\n", args.WorkDir)

	//test
	args = gpu.JobArgs{
		JobName:         "matrix",
		Output:          "123.out",
		Error:           "123.err",
		NumProcess:      1,
		NumTasksPerNode: 1,
		CpusPerTask:     6,
		GpuResources:    "gpu:1",
		CompileScripts:  "module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0;nvcc cuda_add.cu -o cuda_add -lcublas",
		RunScripts:      "module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0;nvcc cuda_add.cu -o cuda_add -lcublas\n./cuda_add",
		Username:        "stu1641",
		Password:        "lj#sJpH4",
		WorkDir:         "Job/gpu",
	}
	server := gpu.NewServer(args, gpu.DefaultJobURL, gpu.DefaultResultURL)
	server.Run()
}
