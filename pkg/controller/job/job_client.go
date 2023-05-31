package job

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// 特别不安全的
const (
	username string = "stu1641"
	password string = "lj#sJpH4"
)

var SiYuan = []string{"64c512g", "a100"}
var PiAndAi = []string{"cpu", "huge", "192c6t", "small", "dgx2"}
var ARM = []string{"arm128c256g"}

func getHost(partition string) (string, error) {
	var host string
	for _, v := range SiYuan {
		if v == partition {
			host = _const.HostSiyuan
			return host, nil
		}
	}
	for _, v := range PiAndAi {
		if v == partition {
			host = _const.HostPiAndAI
			return host, nil
		}
	}
	for _, v := range ARM {
		if v == partition {
			host = _const.HostARM
			return host, nil
		}
	}
	return "", errors.New("[Job Controller] illegal partition")
}

func getAccount(partition string) *object.Account {
	host, err := getHost(strings.ToLower(partition))
	if err != nil {
		fmt.Println(err)
	}
	var account *object.Account
	account = object.NewAccountWith2Para(username, password)
	err = account.SetRemoteBasePath(host)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return account
}

func addJobPod(pod *object.Pod) error {
	suffix := _const.POD_CONFIG_PREFIX + "/" + pod.Name
	podRawData, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRawData)

	req, err := http.NewRequest("PUT", _const.BASE_URI+suffix, reqBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("[ReplicaSet Controller] StatusCode not 200")
	}
	return nil
}
