package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	RUNNING = "RUN"
	FINISH  = "FIN"
)

type WorkflowManager struct {
	ResMap         map[string]string //存放每一个function的result
	StatusMap      *_map.ConcurrentMap
	FunctionMap    *_map.ConcurrentMap    //存放名字到function的映射
	Queue          *queue.ConcurrentQueue //主队列
	Workflow       *object.WorkFlow
	sharedWorkDir  string //共享文件夹的名称
	resChan        chan string
	notifyChan     chan bool
	nowLevelNumber int //当前层有多少函数
}

type FuncEntity struct {
	Parent   string
	Function *object.Function
}

func NewWorkflowManager(workflow *object.WorkFlow, channel chan string) *WorkflowManager {
	resMap := make(map[string]string, 10)
	notiChan := make(chan bool, 10)
	q := &queue.ConcurrentQueue{}
	workflowManger := &WorkflowManager{
		ResMap:         resMap,
		Queue:          q,
		Workflow:       workflow,
		sharedWorkDir:  _const.SHARED_DATA_DIR + "/" + workflow.Name,
		resChan:        channel,
		notifyChan:     notiChan,
		StatusMap:      _map.NewConcurrentMap(),
		FunctionMap:    _map.NewConcurrentMap(),
		nowLevelNumber: 0,
	}
	for _, each := range workflow.FunctionList {
		workflowManger.FunctionMap.Put(each.Name, each)
	}

	return workflowManger
}

func (m *WorkflowManager) Run() {
	go m.syncLoop(m.resChan)
	err := m.callFunctions()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (m *WorkflowManager) syncLoop(resChan <-chan string) {
	//有工作执行完成
	for {
		select {
		case funcName, ok := <-resChan:
			if !ok {
				return
			}
			fmt.Printf("[Workflow] function %s finish!", funcName)
			m.StatusMap.Put(funcName, FINISH)
			if m.isFinish() { //当前层全部函数执行完，通知主线程
				m.notifyChan <- true
			}
		}
	}
}

func (m *WorkflowManager) callFunctions() error {
	funcList := m.Workflow.FunctionList
	var flag = false
	//层序遍历
	for i := 1; i <= m.Workflow.MaxLevel; i++ {
		//选出所有第一层的结点
		if !flag {
			for _, function := range funcList {
				if function.Level == 1 {
					funcEntity := &FuncEntity{
						Function: &function,
						Parent:   "",
					}
					m.Queue.Enqueue(funcEntity)
					m.nowLevelNumber++
				}
			}
		}
		flag = true

		//执行当前层所有函数
		err := m.doBatch()
		if err != nil {
			return err
		}

		//等待，直到被通知
		<-m.notifyChan
		m.nowLevelNumber = 0

		//当前层结点所有子节点入队，且准备输入的参数
		funcNameList := m.StatusMap.GetAllKey()
		for _, each := range funcNameList {
			for _, child := range m.getChild(each) {
				if child == "nil" {
					continue
				} else {
					funcEntity := &FuncEntity{
						Function: m.FunctionMap.Get(child).(*object.Function),
						Parent:   "",
					}
					m.Queue.Enqueue(funcEntity)
					m.nowLevelNumber++
				}
			}

		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *WorkflowManager) doBatch() error {
	for {
		if !m.Queue.Empty() {
			function := m.Queue.Front().(*FuncEntity)
			m.Queue.Dequeue()
			err := m.invokeFunction(function) //可以在前面加go，实现高并发）
			if err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

// 和主函数里的流程有些不同，不需要把函数持久化到数据库中，且共享文件夹目录不同
func (m *WorkflowManager) invokeFunction(funcEntity *FuncEntity) error {
	function := funcEntity.Function
	//先将函数本体上传至etcd
	funcRaw, err := json.Marshal(function)
	if err != nil {
		fmt.Println(err)
		return err
	}
	reqBody := bytes.NewBuffer(funcRaw)
	suffix := _const.FUNC_CONFIG_PREFIX + "/" + function.Name

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//再生成函数的meta数据
	uuid := uuid2.New().String()
	name := function.Name + "_" + uuid
	m.StatusMap.Put(name, RUNNING) //在这里写statusMap

	parent := funcEntity.Parent

	var argList []string
	argList = append(argList, m.getArg(parent))

	meta := object.FunctionMeta{
		Name:    name,
		Type:    object.WORKFLOW,
		ArgList: argList,
	}

	metaRaw, err := json.Marshal(meta)
	if err != nil {
		fmt.Println(err)
		return err
	}
	reqBody2 := bytes.NewBuffer(metaRaw)
	suffix2 := _const.FUNC_RUNTIME_PREFIX + "/" + meta.Name

	req2, err := http.NewRequest("PUT", _const.BASE_URI+suffix2, reqBody2)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *WorkflowManager) isFinish() bool {
	list := m.StatusMap.GetAll()
	for _, each := range list {
		each := each.(string)
		if each == RUNNING {
			return false
		}
	}
	return true
}

// 获取arg
func (m *WorkflowManager) getArg(name string) string {
	//从文件中读取返回值
	fi, err := os.Open(path.Join(m.sharedWorkDir, name))
	if err != nil {
		return ""
	}
	r := bufio.NewReader(fi)
	var flag = false
	var arg string
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil && err != io.EOF {
			return ""
		}
		if err == io.EOF {
			break
		}
		if flag && line != "None" {
			arg = line
			fmt.Println("[Workflow] the arg is:", line)
			return arg
		}
		if line == "the return value is:" {
			flag = true
		}
	}
	return ""
}

func (m *WorkflowManager) getChild(name string) []string {
	nameList := strings.Split(name, "_")
	realName := nameList[0]

	graph := m.Workflow.Graph
	for _, node := range graph {
		if node.Name == realName {
			return node.Child
		}
	}
	return nil
}
