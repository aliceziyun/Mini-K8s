package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/fass-server/fass-util"
	workflow "Mini-K8s/pkg/fass-server/workflow"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type FassServer struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
	queue       queue.ConcurrentQueue
	hashMap     *_map.ConcurrentMap //维护函数和其调用时间
	mtx         sync.Mutex
	resChannel  chan string
}

func NewServer(lsConfig *listwatcher.Config) *FassServer {
	println("[FassServer] fassServer create")

	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[FassServer] list watch start fail...")
	}

	server := &FassServer{
		ls:      ls,
		hashMap: _map.NewConcurrentMap(),
	}
	server.stopChannel = make(chan struct{})
	server.resChannel = nil
	return server
}

func (s *FassServer) Run() {
	fmt.Printf("[FassServer] start running\n")
	go s.register()
	go s.worker()
	//go s.monitor()
	select {}
}

func (s *FassServer) register() {
	go func() {
		err := s.ls.Watch(_const.FUNC_RUNTIME_PREFIX, s.watchNewFunc, s.stopChannel)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("[FassServer] listen fail...\n")
		}
	}()

	go func() {
		err := s.ls.Watch(_const.WORKFLOW_CONFIG_PREFIX, s.watchNewWorkflow, s.stopChannel)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("[FassServer] listen fail...\n")
		}
	}()

}

func (s *FassServer) worker() {
	fmt.Printf("[FassServer] Starting...\n")
	for {
		if !s.queue.Empty() {
			meta := s.queue.Front()
			s.queue.Dequeue()
			err := fass_util.InvokeFunction(meta.(*object.FunctionMeta), s.ls)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (s *FassServer) monitor() { //定时检查有没有需要删除的函数
	for {
		fmt.Printf("[FassServer] monitor...\n")
		for _, each := range s.hashMap.GetAllKey() {
			if time.Now().Sub(s.hashMap.Get(each).(time.Time)).Seconds() > 30 {
				suffix := _const.FUNC_CONFIG_PREFIX + "/" + each
				req, err := http.NewRequest("DELETE", _const.BASE_URI+suffix, nil)
				if err != nil {
					fmt.Println(err)
					return
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Println(err)
					return
				}

				if resp.StatusCode == 200 {
					fmt.Printf("[FassServer] delete function %s \n", each)
					s.hashMap.Remove(each)
				}
			}
		}
		time.Sleep(20 * time.Second) //20秒检查一下
	}
}

func (s *FassServer) watchNewFunc(res etcdstorage.WatchRes) {
	//说明有任务做完了
	if res.ResType == etcdstorage.DELETE {
		fields := strings.Split(res.Key, "/")
		funcName := fields[len(fields)-1]

		//维护一个Ctime的Map
		realName := strings.Split(funcName, "_")[0]

		if realName == "GPU" { //job
			jobName := "job-" + strings.Split(funcName, "_")[1]
			fmt.Printf("[FassServer] job %s finish! Please go to the shared directory to check the result! \n", jobName)

			//更新job的状态
			s.updateJobStatus(jobName)
			return
		}

		fmt.Printf("[FassServer] function %s finish! Please go to the shared directory to check the result! \n", funcName)
		s.hashMap.Put(realName, time.Now())

		if s.resChannel != nil {
			s.resChannel <- funcName
		}
		//这里可以把代码文件删了，不过也没必要
		return
	}

	meta := &object.FunctionMeta{}
	err := json.Unmarshal(res.ValueBytes, &meta)
	if err != nil {
		fmt.Println(err)
	}

	s.queue.Enqueue(meta)
}

func (s *FassServer) watchNewWorkflow(res etcdstorage.WatchRes) {
	fmt.Println("[Workflow] new workflow")
	wf := &object.WorkFlow{}
	err := json.Unmarshal(res.ValueBytes, &wf)
	if err != nil {
		fmt.Println(err)
	}
	//新建一个workflow manager，由manager管理workflow
	s.resChannel = make(chan string)
	go workflow.NewWorkflowManager(wf, s.resChannel).Run()
}

func (s *FassServer) updateJobStatus(name string) {
	res, _ := s.ls.List(_const.JOB_CONFIG_PREFIX + "/" + name)
	job := &object.GPUJob{}
	_ = json.Unmarshal(res[0].ValueBytes, job)

	job.Status = object.FIN
	jobRaw, _ := json.Marshal(job)
	reqBody := bytes.NewBuffer(jobRaw)
	suffix := _const.JOB_CONFIG_PREFIX + "/" + name

	req, _ := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	_, _ = http.DefaultClient.Do(req)

	fmt.Printf("[FassServer] update job %s status", name)
}
