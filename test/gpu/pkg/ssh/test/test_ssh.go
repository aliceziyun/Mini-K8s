package main

import (
	"Mini-K8s/test/gpu/pkg/ssh"
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/melbahja/goph"
)

const (
	gpuUser      = "stu1641"
	gpuPasswd    = "lj#sJpH4"
	gpuLoginAddr = "pilogin.hpc.sjtu.edu.cn"
	gpuDataAddr  = "data.hpc.sjtu.edu.cn"
	accountType  = "acct-stu"

	remoteWorkDir = "Job/gpu/"
	testFilePath  = "Matrix2.cu"
	// remoteWorkDir = "Job"
	// remoteWorkDir = "Job/Matrix.cu"

	jobName = "matrix"
)

// 可以ssh连到pi并且执行命令（且获取命令返回）
func TestSSH() {
	cli, err := goph.NewUnknown(gpuUser, gpuLoginAddr, goph.Password(gpuPasswd))
	defer cli.Close()
	if err != nil {
		fmt.Println("err!")
		log.Fatal(err)
	}

	resp, err := cli.Run("sinfo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got resp:\n %s\n", string(resp))
}

func TestRmdir() {
	cli := ssh.NewClient(gpuUser, gpuPasswd)
	defer cli.Close()
	existed, err1 := cli.ExistsDir(remoteWorkDir)
	// existed, err1 := cli.ExistsFile(testFilePath)

	if err1 != nil {
		fmt.Println(existed)
		fmt.Println("---")
		fmt.Println(err1)
	} else {
		fmt.Println("it is existent?")
		fmt.Println(existed)
		if existed == false {
			fmt.Println("This dir don't exists! No need to remove it.")
			return
		}
	}
	//只能创建下一级的目录，得一级一级建
	resp, err := cli.RmDir(remoteWorkDir)
	if err != nil {
		fmt.Println(resp)
		fmt.Println("sth is wrong")
		fmt.Println(err)
		return
	} else {
		fmt.Println("Successfully rmdir!")
	}
	fmt.Println(resp)
}
func TestMkdir() {
	cli := ssh.NewClient(gpuUser, gpuPasswd)
	defer cli.Close()
	existed, err1 := cli.ExistsDir(remoteWorkDir)
	// existed, err1 := cli.ExistsFile(testFilePath)

	if err1 != nil {
		fmt.Println(existed)
		fmt.Println("---")
		fmt.Println(err1)
	} else {
		fmt.Println("it is existent?")
		fmt.Println(existed)
		if existed == true {
			fmt.Println("This dir already exists! No need to make one.")
			return
		}
	}
	//只能创建下一级的目录，得一级一级建
	resp, err := cli.Mkdir(remoteWorkDir)
	if err != nil {
		fmt.Println(resp)
		fmt.Println("sth is wrong")
		fmt.Println(err)
		return
	} else {
		fmt.Println("Successfully mkdir!")
	}
	fmt.Println(resp)
}

func TestGpuSSH() {
	cli := ssh.NewClient(gpuUser, gpuPasswd)
	// defer cli.Close()
	// cli.CD("Job")

	resp, err := cli.LS()
	if err == nil {
		fmt.Println(resp)
	}
	//job := cli.GetJobById("13277555")
	//fmt.Println(job)
	// fmt.Println(cli.JobCompleted("13462525"))
	//allQueues := cli.GetAllQueueInfo()
	//fmt.Println(allQueues)
	//smallQueue := cli.GetQueueInfoByPartition("small")
	//fmt.Println(smallQueue)
	//fmt.Println(cli.WriteFile("test.txt", "hello world!"))
	//fmt.Println(cli.ReadFile("test.txt"))
	//fmt.Println(cli.CreateFile("test2.txt"))
	//fmt.Println(cli.Mkdir("./test223"))
}

func TestLocalCmd() {
	// command := exec.Command("pwd")

	command := exec.Command("scp", "-r", "/home/data/gpu/", "stu1641@pilogin.hpc.sjtu.edu.cn:/lustre/home/acct-stu/stu1641/Job/")
	//给标准输入以及标准错误初始化一个buffer，每条命令的输出位置可能是不一样的，
	//比如有的命令会将输出放到stdout，有的放到stderr
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}
	//执行命令，直到命令结束
	err := command.Run()
	if err != nil {
		//打印程序中的错误以及命令行标准错误中的输出
		fmt.Println(err)
		fmt.Println(command.Stderr.(*bytes.Buffer).String())
		return
	}
	//打印命令行的标准输出
	fmt.Println(command.Stdout.(*bytes.Buffer).String())
}
func TestCDAndSubmit() {
	cli := ssh.NewClient(gpuUser, gpuPasswd)
	defer cli.Close()

	resp, err := cli.CDAndSbatch("Job", "cuda_add.slurm")
	if err == nil {
		fmt.Println(resp)
	}
	// cli.CD("Job")

}
func main() {
	TestCDAndSubmit()
	// TestSSH()
	// TestMkdir()
	// TestRmdir()
	// TestGpuSSH()
	// TestLocalCmd()
	// fmt.Println(path.Join(remoteWorkDir, jobName+".slurm"))
}
