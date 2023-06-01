package job

import (
	_const "Mini-K8s/cmd/const"
	controller_context "Mini-K8s/cmd/minik8s/controller/controller-context"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/file"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"
)

type JobController struct {
	ls          *listwatcher.ListWatcher
	stopChannel chan struct{}
}

func NewJobController(controllerCtx controller_context.ControllerContext) *JobController {
	jc := &JobController{
		ls:          controllerCtx.Ls,
		stopChannel: make(chan struct{}),
	}

	return jc
}

func (jc *JobController) Run(ctx context.Context) {
	fmt.Println("[Job Controller] start run...")
	jc.register()
	<-ctx.Done()
	close(jc.stopChannel)
}

func (jc *JobController) register() {
	// register job handler
	go func() {
		for {
			err := jc.ls.Watch(_const.JOB_CONFIG_PREFIX, jc.handleJob, jc.stopChannel)
			if err != nil {
				fmt.Println("[Job Controller] list watch RS handler init fail...")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		fmt.Println("[Kubelet] start watch shared data...")
		err := jc.ls.Watch(_const.SHARED_DATA_PREFIX, jc.watchSharedData, jc.stopChannel)
		if err != nil {
			fmt.Printf("[Kubelet] watch shared_data error " + err.Error())
		} else {
			fmt.Println("[Kubelet] return...")
			return
		}
		time.Sleep(1 * time.Second)
	}()
}

func (jc *JobController) handleJob(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}
	job := object.GPUJob{}
	err := json.Unmarshal(res.ValueBytes, &job)
	if err != nil {
		fmt.Println(err)
		return
	}

	account := getAccount(job.Spec.SlurmConfig.Partition)

	//初始化function
	function := &object.Function{
		Name:     "GPU",
		Kind:     "Function",
		Type:     "python",
		FuncName: "GPUjob",
		Path:     "/home/lcz/go/src/Mini-K8s/build/Serveless/GPUjob.py",
		ArgNum:   5,
	}
	funcRaw, err := json.Marshal(function)
	if err != nil {
		fmt.Println(err)
		return
	}
	reqBody := bytes.NewBuffer(funcRaw)
	suffix := _const.FUNC_CONFIG_PREFIX + "/" + function.Name

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	pathName := "job-" + job.Metadata.Uid

	var arglist []string
	arglist = append(arglist, account.GetUsername(), account.GetPassword(), account.GetHost(), "./", ".")

	//初始化function mata
	funcMeta := &object.FunctionMeta{
		Name:    "GPU" + "_" + job.Metadata.Uid,
		ArgList: arglist,
		Type:    "JOB",
		Path:    pathName,
	}

	meta, err := json.Marshal(funcMeta)
	if err != nil {
		fmt.Println(err)
		return
	}

	reqBody2 := bytes.NewBuffer(meta)
	suffix2 := _const.FUNC_RUNTIME_PREFIX + "/" + funcMeta.Name

	req2, err := http.NewRequest("PUT", _const.BASE_URI+suffix2, reqBody2)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[Job Controller] send request to server with code %d", resp2.StatusCode)

}

func (jc *JobController) watchSharedData(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.PUT {
		appZipFile := object.JobAppFile{}
		err := json.Unmarshal(res.ValueBytes, &appZipFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		zipName := appZipFile.Key + ".zip"
		unzippedDir := path.Join(_const.SHARED_DATA_DIR, appZipFile.Key)
		err = file.Bytes2File(appZipFile.App, zipName, _const.SHARED_DATA_DIR)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = file.Unzip(path.Join(_const.SHARED_DATA_DIR, zipName), unzippedDir)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = file.Bytes2File(appZipFile.Slurm, "test.slurm", unzippedDir)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
