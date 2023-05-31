package nodeManager

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/util"
	"encoding/json"
	"fmt"
)

// 恢复nodeManager，核心是告诉自己是哪个node
func (m *NodeManager) recover() bool {
	resList, err := m.ls.List(_const.NODE_CONFIG_PREFIX)
	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, res := range resList {
		result := &object.Node{}
		err = json.Unmarshal(res.ValueBytes, result)
		if result.Spec.DynamicIp == util.GetDynamicIP() { //是自己的node
			m.NodeName = result.MetaData.Name
			m.DynamicIp = result.Spec.DynamicIp
			m.node = result
			_const.NODE_NAME = m.NodeName
			fmt.Printf("[Node Manager] recover with %s name", _const.NODE_NAME)
			return true
		}
	}

	return false
}
