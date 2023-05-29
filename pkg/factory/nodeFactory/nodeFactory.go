package nodeFactory

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

var instance *NetMap

type NetMap struct {
	Name2Node      map[string]*object.Node
	BasicIpAndPair string
}

func NewNode(node *object.Node) (*object.Node, error) {
	//将node加入映射中
	var rwLock sync.RWMutex
	rwLock.RLock()
	defer rwLock.RUnlock()
	netMap := getNetMap()
	if netMap.isExist(node.Spec.DynamicIp) {
		return nil, errors.New(node.Spec.DynamicIp + "already exist")
	}
	var name string
	name = node.MetaData.Name //默认有名字
	node.MetaData.Name = name
	node.MetaData.Uid = uuid.New().String()
	netMap.Name2Node[name] = node
	fmt.Println("new node added")
	return node, nil
}

func getNetMap() *NetMap {
	if instance == nil {
		instance := &NetMap{}
		instance.Name2Node = make(map[string]*object.Node)
		instance.BasicIpAndPair = _const.BASIC_IP_AND_MASK
		return instance
	} else {
		return instance
	}
}

func (netMap *NetMap) isExist(ip string) bool {
	for _, v := range netMap.Name2Node {
		if v.Spec.DynamicIp == ip {
			return true
		}
	}
	return false
}
