package nodeManager

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	dynamicIp    string //浮动ip
	nodeName     string //节点的名称
	ipAndMask    string //节点的docker网段
	err          error
	reboot       bool
}

type NodeSnapShot struct {
	DynamicIp string
	IpAndMask string
	NodeName  string
	Error     string
}

func (m *NodeManager) StartKubeNetSupport() error {
	fmt.Println("[Node Manager] manager start...")
	if m.reboot {
		return nil
	}
	return m.registerNode()
}

func (m *NodeManager) register() {
	go func() {
		for {
			err := m.ls.Watch(_const.NODE_CONFIG_PREFIX, m.watchNode, m.stopChannel)
			if err != nil {
				fmt.Println("[Node Manager] watch register error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}()
}

func (m *NodeManager) registerNode() error {
	m.register()
	fmt.Println("[Node Manager] start init...")

	//boot.BootFlannel()

	//发起注册的http请求
	suffix := _const.NODE_CONFIG_PREFIX + "/" + m.dynamicIp
	var node *object.Node
	if m.node == nil {
		node = &object.Node{
			MasterIp: _const.MASTER_IP,
			Spec: object.NodeSpec{
				DynamicIp:     m.dynamicIp,
				NodeIpAndMask: m.ipAndMask,
			},
		}
	} else {
		node = &object.Node{
			MetaData: m.node.MetaData,
			MasterIp: m.node.MasterIp,
			Spec: object.NodeSpec{
				DynamicIp:     m.dynamicIp,
				NodeIpAndMask: m.ipAndMask,
			},
		}
	}
	err := m.putNode(suffix, node)
	if err != nil {
		return err
	}
	return nil
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
	if node.Spec.DynamicIp == m.dynamicIp {
		m.nodeName = node.MetaData.Name
	}
}
