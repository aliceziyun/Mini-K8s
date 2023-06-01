package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/file"
	_map "Mini-K8s/third_party/map"
	"Mini-K8s/third_party/queue"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	RUNNING = "RUN"
	FINISH  = "FIN"
	REMOVED = "REMOVED"
)

type WorkflowManager struct {
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
	Function object.Function
}

func NewWorkflowManager(workflow *object.WorkFlow, channel chan string) *WorkflowManager {
	notiChan := make(chan bool, 10)
	q := &queue.ConcurrentQueue{}
	workflowManger := &WorkflowManager{
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
		workflowManger.FunctionMap.Put(each.FuncName, each)
	}

	return workflowManger
}

func (m *WorkflowManager) Run() {
	fmt.Println("[Workflow] Run...")
	go m.syncLoop(m.resChan)
	err := m.callFunctions()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (m *WorkflowManager) syncLoop(resChan <-chan string) {
	fmt.Println("[Workflow] start syncLoop...")
	//有工作执行完成
	for {
		select {
		case funcName, ok := <-resChan:
			if !ok {
				return
			}
			fmt.Printf("[Workflow] function %s finish!", funcName)
			m.StatusMap.Put(funcName, FINISH)
			m.nowLevelNumber--
			if m.isFinish() { //当前层全部函数执行完，通知主线程
				fmt.Println("[Workflow] clear")
				m.notifyChan <- true
			}
		}
	}
}

func (m *WorkflowManager) callFunctions() error {
	fmt.Println("[Workflow] start calling function...")
	funcList := m.Workflow.FunctionList
	var flag = false
	//层序遍历
	for i := 1; i <= m.Workflow.MaxLevel; i++ {
		//选出所有第一层的结点
		if !flag {
			for _, function := range funcList {
				if function.Level == 1 {
					funcEntity := &FuncEntity{
						Function: function,
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
			if m.StatusMap.Get(each).(string) == REMOVED {
				continue
			}
			for _, child := range m.getChild(each) {
				if child == "nil" {
					continue
				} else {
					funcEntity := &FuncEntity{
						Function: m.FunctionMap.Get(child).(object.Function),
						Parent:   each,
					}
					m.Queue.Enqueue(funcEntity)
					m.nowLevelNumber++
				}
			}
			m.StatusMap.Put(each, REMOVED)
		}
		if err != nil {
			return err
		}
	}
	//结束之后收集所有的结果
	m.collectResult()

	return nil
}

func (m *WorkflowManager) doBatch() error {
	fmt.Println("[Workflow] start batching...")
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
	function.Name = function.FuncName
	funcRaw, err := json.Marshal(function)
	if err != nil {
		fmt.Println(err)
		return err
	}
	reqBody := bytes.NewBuffer(funcRaw)
	suffix := _const.FUNC_CONFIG_PREFIX + "/" + function.FuncName

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
	name := function.FuncName + "_" + uuid
	m.StatusMap.Put(name, RUNNING) //在这里写statusMap

	parent := funcEntity.Parent
	var argList []string

	if parent == "" { //初始节点，读取yaml中的参数
		argList = m.Workflow.InitialArgs
	} else {
		argList = append(argList, m.getArg(parent))
	}

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
	if m.nowLevelNumber == 0 {
		return true
	}
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
	fmt.Println("[Workflow] get args")
	//从文件中读取返回值
	//fi, err := os.Open(path.Join(m.sharedWorkDir, name))
	//fmt.Println(path.Join(m.sharedWorkDir, name))
	fi, err := os.Open(path.Join(_const.SHARED_DATA_DIR, name, "output.txt"))
	if err != nil {
		return ""
	}
	r := bufio.NewReader(fi)
	var arg string
	var upLine string
	for {
		upLine = arg
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		arg = line
		if err != nil && err != io.EOF {
			return ""
		}
		if err == io.EOF {
			break
		}
	}
	if upLine == "None" {
		return ""
	} else {
		return upLine
	}
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

func (m *WorkflowManager) collectResult() {
	var res = ""
	for _, each := range m.StatusMap.GetAllKey() {
		dataPath := path.Join(_const.SHARED_DATA_DIR, each)
		data, _ := ioutil.ReadFile(path.Join(dataPath, "output.txt"))
		res = res + each + ":\n" + string(data) + "\n"
	}
	fmt.Println(res)
	raw := []byte(res)
	err := file.Bytes2File(raw, m.Workflow.Name, path.Join(_const.SHARED_DATA_DIR, m.Workflow.Name))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("[Workflow] workflow %s finish", m.Workflow.Name)
}
