package ssh

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/melbahja/goph"
	"github.com/spf13/cast"
)

const (
	gpuUser      = "stu1641"
	gpuPasswd    = "lj#sJpH4"
	gpuLoginAddr = "pilogin.hpc.sjtu.edu.cn"
	// gpuDataAddr  = "data.hpc.sjtu.edu.cn"
	gpuDataAddr = "pilogin.hpc.sjtu.edu.cn"
	accountType = "acct-stu"
)

type JobInfo struct {
	JobID     string
	JobName   string
	Partition string
	Account   string
	AllocCPUS int
	State     string
	ExitCode  string
}

type QueueInfo struct {
	Partition string
	Available string
	TimeLimit string
	Nodes     int
	State     string
	NodeList  string
}

type Client interface {
	Close()
	Reconnect()

	//remote的命令
	GetQueueInfoByPartition(partition string) []*QueueInfo
	GetAllQueueInfo() []*QueueInfo                                   //Sinfo
	GetJobById(jobID string) *JobInfo                                //Squeue
	SubmitJob1(scriptPath string, scriptName string) (string, error) //Sbatch
	// SubmitJob(scriptPath string) (string, error)                     //Sbatch
	JobCompleted(jobID string) bool
	//Scancel() ([]byte, error)         //取消指定作业

	// Compile(cmd string) (string, error)

	//本地命令
	ScpFromLocal(localPath, remotePath string) error
	ScpToLocal(localPath, remotePath string) error //scp也可以写一个远端操作的
	Rsync(localPath, remotePath string) error      //Rsync必须在本地操作
	MkLocalDir(dir string) error
	TryMkLocalDir(dir string) error
	ClearLocalDir(dir string) error //清除(本地)目录下所有文件

	//remote的命令
	LS() (string, error)
	PWD() (string, error)
	CD(filepath string) (string, error)                           //单独CD好像没啥用，因为每次会新起一个sshSession...
	CDAndSbatch(filepath string, fileName string) (string, error) //需要保证filepath正确(?)
	ExistsFile(filepath string) (bool, error)
	ExistsDir(filepath string) (bool, error)
	Mkdir(dir string) (string, error)
	TryMkdir(dir string) (string, error)
	RmDir(dir string) (string, error)
	TryRmDir(dir string) (string, error)
	CreateFile(filename string) (string, error)
	WriteFile(filename, content string) (string, error)
	ReadFile(filename string) (string, error)
}

type client struct {
	username string
	password string
	sshCli   *goph.Client
}

func (cli *client) JobCompleted(jobID string) bool {
	job := cli.GetJobById(jobID)
	return job != nil && job.State == "COMPLETED"
}

func (cli *client) Reconnect() {
	if cli.sshCli != nil {
		_ = cli.sshCli.Close()
	}

	cli.sshCli = newSSHClient(cli.username, cli.password)
}

func (cli *client) ExistsFile(filepath string) (bool, error) {
	cmd := fmt.Sprintf("if [ -f %s ]; then echo true; else echo false; fi", filepath)
	raw, err := cli.sshCli.Run(cmd)
	if err != nil {
		return false, err
	}
	// fmt.Printf("the res is : %s\n", raw)
	// fmt.Println(raw)
	// fmt.Println(string(raw))
	tmpStr := strings.Replace(string(raw), "\n", "", -1)
	// fmt.Println(tmpStr)
	// fmt.Println(cast.ToBool(tmpStr))
	// fmt.Println(cast.ToBool(string("true\n")))
	// return cast.ToBool(string(raw)), nil
	return cast.ToBool(tmpStr), nil
}

func (cli *client) ExistsDir(filepath string) (bool, error) {
	cmd := fmt.Sprintf("if [ -e %s ]; then echo true; else echo false; fi", filepath)
	raw, err := cli.sshCli.Run(cmd)
	if err != nil {
		return false, err
	}
	tmpStr := strings.Replace(string(raw), "\n", "", -1)

	// return cast.ToBool(string(raw)), nil
	return cast.ToBool(tmpStr), nil
}

// func (cli *client) Compile(cmd string) (string, error) {
// 	//if resp, err := cli.loadCuda(); err != nil {
// 	//	return resp, err
// 	//}
// 	fmt.Println(cmd)
// 	resp, err := cli.sshCli.Run(cmd)
// 	return string(resp), err
// }

func (cli *client) ScpFromLocal(localPath, remotePath string) error {
	if runtime.GOOS == "linux" {
		remoteAddr := fmt.Sprintf("%s@%s:%s", cli.username, gpuDataAddr, remotePath)
		cmd := exec.Command("scp", "-r", localPath, remoteAddr)
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
		}
		//打印命令行的标准输出
		fmt.Println(cmd.Stdout.(*bytes.Buffer).String())
		return err
	}
	return fmt.Errorf("scp is not supported in your os")
}

func (cli *client) ScpToLocal(localPath, remotePath string) error {
	if runtime.GOOS == "linux" {
		remoteAddr := fmt.Sprintf("%s@%s:%s", cli.username, gpuDataAddr, remotePath)
		cmd := exec.Command("scp", remoteAddr, localPath)
		return cmd.Run()
	}
	return fmt.Errorf("scp is not supported in your os")
}

// 从远端同步文件夹到本地
// local = "/home/data/gpu/"
// remote = "/lustre/home/acct-stu/stu1641/Job/gpu/"
func (cli *client) Rsync(localPath, remotePath string) error {
	if runtime.GOOS == "linux" {
		remoteAddr := fmt.Sprintf("%s@%s:%s", cli.username, gpuDataAddr, remotePath)
		cmd := exec.Command("rsync", "--archive", "--partial", "--progress", remoteAddr, localPath)
		return cmd.Run()
	}
	return fmt.Errorf("rsync is not supported in your os")
}

// func (cli *client) loadCuda() (string, error) {
// 	cmd := "module load cuda/9.2.88-gcc-4.8.5"
// 	resp, err := cli.sshCli.Run(cmd)
// 	return string(resp), err
// }

func (cli *client) SubmitJob1(scriptPath string, scriptName string) (string, error) {
	//初始时是/stu1641目录
	// resp0, err0 := cli.PWD()
	// if err0 != nil {
	// 	fmt.Println(err0)
	// }
	// fmt.Println(resp0)

	//cd到工作目录, 然后交作业
	resp, err0 := cli.CDAndSbatch(scriptPath, scriptName)
	if err0 != nil {
		fmt.Println(err0)
	}
	var jobID string
	fmt.Printf("Submitted batch job and got response: %s\n", resp)
	n, err := fmt.Sscanf(resp, "Submitted batch job %s", &jobID)
	if err != nil || n != 1 {
		fmt.Println("maybe scriptPatch is wrong?")
		return "-1", err
	}
	return jobID, nil
}

func (cli *client) TryMkLocalDir(dir string) error {
	if runtime.GOOS == "linux" {
		existed, err1 := cli.ExistsDir(dir)
		if err1 != nil {
			fmt.Println(err1)
		} else {
			if existed == true {
				fmt.Println("This dir already exists! No need to make one.")
				return nil
			}
		}
		err := cli.MkLocalDir(dir)
		return err
	}
	return fmt.Errorf("mkdir is not supported in your os")
}

func (cli *client) MkLocalDir(dir string) error {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("mkdir", dir)
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
		}
		fmt.Println(cmd.Stdout.(*bytes.Buffer).String())
		return err
	}
	return fmt.Errorf("mkdir is not supported in your os")
}

// func (cli *client) SubmitJob(scriptPath string) (string, error) {
// 	cmd := fmt.Sprintf("sbatch %s", scriptPath)
// 	respRaw, err := cli.sshCli.Run(cmd)
// 	resp := string(respRaw)
// 	var jobID string
// 	fmt.Printf("Submit and got response: %s\n", resp)
// n, err := fmt.Sscanf(resp, "Submitted batch job %s", &jobID)
// 	if err != nil || n != 1 {
// 		return "-1", err
// 	}
// 	return jobID, nil
// }

// sshCli.Run是新起一个sshSession，所以单独cd似乎没有实质作用
func (cli *client) CD(filepath string) (string, error) {
	cmd := fmt.Sprintf("cd %s", filepath)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) CDAndSbatch(filepath string, fileName string) (string, error) {
	cmd := fmt.Sprintf(
		"cd %s\nsbatch %s",
		filepath,
		fileName,
	)
	// cmd := fmt.Sprintf("sbatch %s", scriptName)
	respRaw, err := cli.sshCli.Run(cmd)
	return string(respRaw), err
}

func (cli *client) LS() (string, error) {
	cmd := fmt.Sprintf("ls")
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) PWD() (string, error) {
	cmd := fmt.Sprintf("pwd")
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) RmDir(dir string) (string, error) {
	cmd := fmt.Sprintf("rm -rf %s", dir)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}
func (cli *client) TryRmDir(dir string) (string, error) {

	existed, err1 := cli.ExistsDir(dir)
	// existed, err1 := cli.ExistsFile(testFilePath)

	if err1 != nil {
		fmt.Println(err1)
	} else {
		if existed == false {
			fmt.Println("This dir don't exists! No need to remove it.")
			return "nothing", nil
		}
	}

	resp, err := cli.RmDir(dir)
	if err != nil {
		fmt.Println(err)
	}
	// else {
	// 	// fmt.Println("Successfully rmdir!")
	// }
	fmt.Println(resp)
	return string(resp), err
}

// remote dir
func (cli *client) Mkdir(dir string) (string, error) {
	cmd := fmt.Sprintf("mkdir %s", dir)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) TryMkdir(dir string) (string, error) {
	existed, err1 := cli.ExistsDir(dir)
	if err1 != nil {
		// fmt.Println(existed)
		// fmt.Println("---")
		fmt.Println(err1)
	} else {
		// fmt.Println("it is existent?")
		// fmt.Println(existed)
		if existed == true {
			fmt.Println("This dir already exists! No need to make one.")
			return "nothing", nil
		}
	}
	resp, err := cli.Mkdir(dir)
	if err != nil {
		fmt.Println(resp)
		// fmt.Println("sth is wrong")
		fmt.Println(err)
	} else {
		// fmt.Println("Successfully mkdir!")
	}
	return string(resp), err
}

func (cli *client) ClearLocalDir(dir string) error {
	if runtime.GOOS == "linux" {
		tmpDir := dir
		// 这里通配符没有效果，不知道为什么。自己开命令行是可以的。。。
		// if dir[len(dir)-1] == '/' {
		// 	tmpDir += "*"
		// } else {
		// 	tmpDir += "/*"
		// 	// tmpDir += "/"
		// }

		//所以这里就先把整个文件夹删了再创一个新的
		cmd := exec.Command("rm", "-rf", tmpDir)
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
			fmt.Println(cmd.Stdout.(*bytes.Buffer).String())
		}
		cmd = exec.Command("mkdir", tmpDir)
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err = cmd.Run()
		// fmt.Println(cmd.String())
		if err != nil {
			fmt.Println(err)
			fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
		}
		fmt.Println(cmd.Stdout.(*bytes.Buffer).String())
		return err
	}
	return fmt.Errorf("rm is not supported in your os")

}

func (cli *client) CreateFile(filename string) (string, error) {
	cmd := fmt.Sprintf("touch %s", filename)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) WriteFile(filename, content string) (string, error) {
	content = strings.Replace(content, "\"", "\\\"", -1)
	fmt.Println(content)
	cmd := fmt.Sprintf("echo \"%s\" > %s", content, filename)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) ReadFile(filename string) (string, error) {
	cmd := fmt.Sprintf("cat %s", filename)
	resp, err := cli.sshCli.Run(cmd)
	return string(resp), err
}

func (cli *client) Close() {
	cli.sshCli.Close()
}

func (cli *client) GetJobById(jobID string) *JobInfo {
	cmd := fmt.Sprintf("sacct -j %s | tail -n +3 | awk '{print $1, $2, $3, $4, $5, $6, $7}'", jobID)
	if raw, err := cli.sshCli.Run(cmd); err == nil {
		resp := string(raw)
		rows := strings.Split(resp, "\n")
		if len(rows) > 0 {
			row := rows[0]
			cols := strings.Split(row, " ")
			if len(cols) == 7 {
				return &JobInfo{
					JobID:     cols[0],
					JobName:   cols[1],
					Partition: cols[2],
					Account:   cols[3],
					AllocCPUS: cast.ToInt(cols[4]),
					State:     cols[5],
					ExitCode:  cols[6],
				}
			}
		}
	}
	return nil
}

func parseQueueInfoTable(raw string) (infos []*QueueInfo) {
	rows := strings.Split(raw, "\n")
	for _, row := range rows {
		cols := strings.Split(row, " ")
		if len(cols) != 6 {
			continue
		}
		infos = append(infos, &QueueInfo{
			Partition: cols[0],
			Available: cols[1],
			TimeLimit: cols[2],
			Nodes:     cast.ToInt(cols[3]),
			State:     cols[4],
			NodeList:  cols[5],
		})
	}
	return
}

func (cli *client) GetAllQueueInfo() (infos []*QueueInfo) {
	cmd := "sinfo | tail -n +2 | awk '{print $1, $2, $3, $4, $5, $6}'"
	if raw, err := cli.sshCli.Run(cmd); err == nil {
		return parseQueueInfoTable(string(raw))
	}
	return nil
}

func (cli *client) GetQueueInfoByPartition(partition string) (infos []*QueueInfo) {
	cmd := fmt.Sprintf("sinfo --partition=%s | tail -n +2 | awk '{print $1, $2, $3, $4, $5, $6}'", partition)
	if raw, err := cli.sshCli.Run(cmd); err == nil {
		return parseQueueInfoTable(string(raw))
	}
	return nil
}

func (cli *client) Scancel() ([]byte, error) {
	return cli.sshCli.Run("scancel")
}

func (cli *client) Upload(localPath, remotePath string) error {
	return cli.sshCli.Upload(localPath, remotePath)
}

// 新建一个sshClient（连到remote）
func newSSHClient(username, password string) *goph.Client {
	if cli, err := goph.NewUnknown(username, gpuLoginAddr, goph.Password(password)); err == nil {
		return cli
	}
	return nil
}

func NewClient(username, password string) Client {
	sshCli := newSSHClient(username, password)
	return &client{
		username: username,
		password: password,
		sshCli:   sshCli,
	}
}
