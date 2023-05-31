package nodeManager

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/util"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type NodeManager struct {
	rwLock       sync.RWMutex
	nodeSnapShot NodeSnapShot
	node         *object.Node
	ls           *listwatcher.ListWatcher
	stopChannel  <-chan struct{}
	DynamicIp    string //浮动ip
	NodeName     string //节点的名称
	IpAndMask    string //节点的docker网段
	err          error
	reboot       bool
}

type NodeSnapShot struct {
	DynamicIp string
	IpAndMask string
	NodeName  string
	Error     string
}

func NewNodeManager(lsConfig *listwatcher.Config) (*NodeManager, error) {
	nodeManager := &NodeManager{}
	var rwLock sync.RWMutex
	nodeManager.rwLock = rwLock
	nodeManager.stopChannel = make(chan struct{}, 10)
	nodeManager.DynamicIp = util.GetDynamicIP()
	//nodeManager.IpAndMask = tools.GetDocker0IpAndMask()
	ls, err2 := listwatcher.NewListWatcher(lsConfig)
	if err2 != nil {
		return nil, err2
	}
	nodeManager.ls = ls

	resNode, _ := nodeManager.getNode(nodeManager.DynamicIp)
	if resNode == nil {
		nodeManager.node = nil
		nodeManager.reboot = false
	} else {
		nodeManager.node = resNode
		nodeManager.reboot = true
		nodeManager.NodeName = resNode.MetaData.Name
	}
	sErr := ""
	if nodeManager.err != nil {
		sErr = nodeManager.err.Error()
	}
	nodeManager.nodeSnapShot = NodeSnapShot{
		IpAndMask: nodeManager.IpAndMask,
		DynamicIp: nodeManager.DynamicIp,
		NodeName:  nodeManager.NodeName,
		Error:     sErr,
	}

	_ = nodeManager.recover()

	return nodeManager, nil
}

func (m *NodeManager) Start() {
	fmt.Println("[Node Manager] manager start...")
	if m.reboot {
		return
	}
	m.registerNode()
}

func (m *NodeManager) register() {
	go func() {
		for {
			err := m.ls.Watch(_const.NODE_CONFIG_PREFIX, m.watchNode, m.stopChannel)
			if err != nil {
				fmt.Println("[Node Manager] watch register error" + err.Error())
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func (m *NodeManager) registerNode() {
	go m.register()
	fmt.Println("[Node Manager] start init...")

	//boot.BootFlannel()

	//发起注册的http请求（没必要注册吧
	//suffix := _const.NODE_CONFIG_PREFIX + "/" + m.DynamicIp
	//var node *object.Node
	//if m.node == nil {
	//	node = &object.Node{
	//		MasterIp: _const.MASTER_IP,
	//		Spec: object.NodeSpec{
	//			DynamicIp:     m.DynamicIp,
	//			NodeIpAndMask: m.IpAndMask,
	//		},
	//	}
	//} else {
	//	node = &object.Node{
	//		MetaData: m.node.MetaData,
	//		MasterIp: m.node.MasterIp,
	//		Spec: object.NodeSpec{
	//			DynamicIp:     m.DynamicIp,
	//			NodeIpAndMask: m.IpAndMask,
	//		},
	//	}
	//}
	//err := m.putNode(suffix, node)
	//if err != nil {
	//	return
	//}
	return
}

func (m *NodeManager) putNode(suffix string, requestBody any) error {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	reqBody := bytes.NewBuffer(body)
	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		return err
	}
	resp, err3 := http.DefaultClient.Do(req)
	if err3 != nil {
		return err3
	}
	if resp.StatusCode != 200 {
		return errors.New("StatusCode not 200")
	}
	return nil
}

func (m *NodeManager) getNode(ip string) (*object.Node, error) {
	suffix := _const.NODE_CONFIG_PREFIX + "/" + ip

	request, err := http.NewRequest("GET", _const.BASE_URI+suffix, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}

	reader := response.Body
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var resList []etcdstorage.ListRes
	err = json.Unmarshal(data, &resList)
	if err != nil {
		return nil, err
	}

	if len(resList) == 0 {
		return nil, nil
	}

	result := &object.Node{}
	err = json.Unmarshal(resList[0].ValueBytes, result)
	return result, err
}

// 只更新本机的nodeName
func (m *NodeManager) watchNode(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	}

	node := &object.Node{}
	err := json.Unmarshal(res.ValueBytes, node)
	if err != nil {
		fmt.Println(err)
		return
	}

	//更新nodeName
	if node.Spec.DynamicIp == m.DynamicIp {
		m.NodeName = node.MetaData.Name
		_const.NODE_NAME = node.MetaData.Name
	}

	fmt.Println("[Node Manager] new node name", m.NodeName)
}
