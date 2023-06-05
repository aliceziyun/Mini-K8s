package gpu

import (
	"Mini-K8s/test/gpu/pkg/ssh"
	"Mini-K8s/test/gpu/pkg/utils/wait"
	"runtime"

	// "Mini-K8s/util/uidutil"

	"fmt"
	"path"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type JobArgs struct {
	JobName         string
	WorkDir         string
	Output          string
	Error           string
	NumProcess      int
	NumTasksPerNode int
	CpusPerTask     int
	GpuResources    string

	RunScripts     string
	CompileScripts string
	Username       string
	Password       string
}

const pollPeriod = time.Second * 10
const DefaultJobURL = "/home/data/gpu"
const DefaultResultURL = "/home/data/res"

type Server interface {
	Run()
}

type server struct {
	cli     ssh.Client
	args    JobArgs
	uid     string
	jobsURL string
	resURL  string
	jobID   string
}

// recoverutil
func Trace(errorMsg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var str strings.Builder
	str.WriteString(errorMsg)
	str.WriteString("Trace back\n")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\tat %s:%d\n\t\t%s\n", file, line, fn.Name()))
	}
	return str.String()
}

func (s *server) recover() {
	if err := recover(); err != nil {
		fmt.Println(Trace(fmt.Sprintf("%v\n", err)))
		s.cli.Reconnect()
	}
}

// 定时查看提交的作业是否completed
func (s *server) checkCompleted() bool {
	defer s.recover()
	fmt.Printf("Checking Job %s state......\n", s.jobID)
	return !s.cli.JobCompleted(s.jobID)
}

// func (s *server) getCudaFiles() []string {
// 	var cudaFiles []string
// 	_ = filepath.WalkDir(s.jobsURL, func(path string, d fs.DirEntry, err error) error {
// 		if !d.IsDir() {
// 			fileName := d.Name()
// 			if strings.HasSuffix(fileName, ".cu") {
// 				cudaFiles = append(cudaFiles, fileName)
// 			}
// 		}
// 		return nil
// 	})
// 	fmt.Printf("cudaFiles: %v\n", cudaFiles)
// 	return cudaFiles
// }

// func (s *server) uploadSmallFiles(filenames []string) error {
// 	if resp, err := s.cli.Mkdir(s.args.WorkDir); err != nil {
// 		fmt.Println(resp)
// 		return err
// 	}
// 	for _, filename := range filenames {
// 		if file, err := os.Open(path.Join(s.jobsURL, filename)); err == nil {
// 			if content, err := ioutil.ReadAll(file); err == nil {
// 				_, _ = s.cli.WriteFile(path.Join(s.args.WorkDir, filename), string(content))
// 			}
// 		}
// 	}
// 	return nil
// }

func (s *server) scriptPath() string {
	// return path.Join(s.args.WorkDir, s.args.JobName+"-"+s.uid+".slurm")
	return path.Join(s.args.WorkDir, s.args.JobName+".slurm")
}

// func (s *server) createJobScript() error {
// 	template := `#!/bin/bash
// #SBATCH --job-name=%s
// #SBATCH --partition=dgx2
// #SBATCH --output=%s
// #SBATCH --error=%s
// #SBATCH -N %d
// #SBATCH --ntasks-per-node=%d
// #SBATCH --cpus-per-task=%d
// #SBATCH --gres=%s

// ulimit -s unlimited
// ulimit -l unlimited

// %s
// `
// 	script := fmt.Sprintf(
// 		template,
// 		s.args.JobName,
// 		s.args.Output,
// 		s.args.Error,
// 		s.args.NumProcess,
// 		s.args.NumTasksPerNode,
// 		s.args.CpusPerTask,
// 		s.args.GpuResources,
// 		strings.Replace(s.args.RunScripts, ";", "\n", -1),
// 	)
// 	_, err := s.cli.WriteFile(s.scriptPath(), script)
// 	return err
// }

// func (s *server) compile() error {
// 	_, err := s.cli.Compile(s.args.CompileScripts)
// 	return err
// }

func (s *server) submitJob() (err error) {
	// if s.jobID, err = s.cli.SubmitJob(s.scriptPath()); err == nil {
	// 	fmt.Printf("submit succeed, got jod ID: %s\n", s.jobID)
	// }
	s.jobID, err = s.cli.SubmitJob1(s.args.WorkDir, s.args.JobName+".slurm")
	if err == nil {
		fmt.Printf("Submit the job succeed. Jod ID = %s\n", s.jobID)
	}
	return err
}

// 从本地上传cuda和slurm文件到remote
func (s *server) uploadCuda() (err error) {
	if resp, err := s.cli.TryMkdir(s.args.WorkDir); err != nil {
		fmt.Println(resp)
		return err
	}
	err = s.cli.ScpFromLocal(
		// "/home/data/gpu/",
		s.jobsURL,
		"/lustre/home/acct-stu/stu1641/Job/",
	)

	return err
}

// 从本地jobsUrl获取cuda文件并将其写到远端文件，然后编译，在远端创建slurm脚本
// func (s *server) prepare() (err error) {
// 	// 从本地jobsUrl获取cuda文件
// 	cudaFiles := s.getCudaFiles()
// 	if len(cudaFiles) == 0 {
// 		return fmt.Errorf("no available cuda files")
// 	}
// 	//将cuda文件写到远端
// 	if err = s.uploadSmallFiles(cudaFiles); err != nil {
// 		return err
// 	}
// 	fmt.Println("upload cuda files successfully")
// 	//编译
// 	if err = s.compile(); err != nil {
// 		return err
// 	}
// 	fmt.Println("compile successfully")
// 	// 在远端创建slurm脚本
// 	if err = s.createJobScript(); err != nil {
// 		return err
// 	}
// 	fmt.Println("create job script successfully")
// 	return nil
// }

func (s *server) syncResult() {
	// s.cli.Rsync("/home/data/gpu/",
	// 	"/lustre/home/acct-stu/stu1641/Job/gpu/")
	// s.cli.ScpToLocal("/home/data/res/res"+s.jobID+"/", "/lustre/home/acct-stu/stu1641/"+s.jobID+".err")
	// s.cli.ScpToLocal("/home/data/res/res"+s.jobID+"/", "/lustre/home/acct-stu/stu1641/"+s.jobID+".out")
	s.cli.ScpToLocal("/home/data/res/res"+s.jobID+"/", "/lustre/home/acct-stu/stu1641/Job/gpu/"+s.jobID+".err")
	s.cli.ScpToLocal("/home/data/res/res"+s.jobID+"/", "/lustre/home/acct-stu/stu1641/Job/gpu/"+s.jobID+".out")
}

// func (s *server) downloadResult() {
// 	outputFile := s.args.Output
// 	if content, err := s.cli.ReadFile(outputFile); err == nil {
// 		if file, err := os.Create(path.Join(s.jobsURL, outputFile)); err == nil {
// 			defer file.Close()
// 			_, _ = file.Write([]byte(content))
// 		}
// 	}

// 	errorFile := s.args.Error
// 	if content, err := s.cli.ReadFile(errorFile); err == nil {
// 		if file, err := os.Create(path.Join(s.jobsURL, errorFile)); err == nil {
// 			defer file.Close()
// 			_, _ = file.Write([]byte(content))
// 		}
// 	}
// }

func (s *server) Run() {
	// var (
	// 	reset = string([]byte{27, 91, 48, 109})
	// 	red   = string([]byte{27, 91, 57, 49, 109})
	// )

	if err := s.uploadCuda(); err != nil {
		// fmt.Println(red + "[Error] " + fmt.Sprintf("[uploadCuda]: "+err.Error()) + reset)
		fmt.Println("[Error]: when uploading Cuda files")
		return
	}

	// // 从本地jobsUrl获取cuda文件并将其写到远端文件，然后编译，在远端创建slurm脚本
	// if err := s.prepare(); err != nil {
	// 	// logger.Error("prepare: " + err.Error())
	// 	fmt.Println(red + "[Error] " + fmt.Sprintf("prepare: "+err.Error()) + reset)
	// 	return
	// }

	// 使用sbatch提交作业，并获取JobID
	if err := s.submitJob(); err != nil {
		// fmt.Println(red + "[Error] " + fmt.Sprintf("submit: "+err.Error()) + reset)
		fmt.Println("[Error]: when submitting the job")
		return
	}

	// s.jobID = "25429622" //先设置一个特定的id，测试查看job是否完成，正常应该是submit的时候会返回id

	// // 等到作业状态为COMPLETED，即表示作业已完成
	wait.PeriodWithCondition(pollPeriod, pollPeriod, s.checkCompleted)

	fmt.Printf("Job %s has been finished, now sync the result to local\n", s.jobID)

	// 从远端将结果文件同步到本地
	s.cli.TryMkLocalDir("/home/data/res/res" + s.jobID)
	s.syncResult()

	fmt.Println("Successfully Sync. Now hang...")
	//清理远端文件
	s.cli.TryRmDir(s.args.WorkDir)

	// wait.Forever()
}

func NewUuid() string {
	return uuid.NewV4().String()
}
func NewServer(args JobArgs, jobsURL string, resURL string) Server {
	return &server{
		cli:     ssh.NewClient(args.Username, args.Password),
		args:    args,
		uid:     NewUuid(),
		jobsURL: jobsURL,
		resURL:  resURL,
	}
}
