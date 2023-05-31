package fass_server

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func getFunctionBody(body string, argNum int, isReturn bool, funcName string, uuid string) []byte {
	body = appendRedirector(body)

	body = body + "    print(" + funcName + "("
	//根据参数数量修饰函数体
	for i := 0; i < argNum; i++ {
		index := fmt.Sprintf("%d", i+1) //因为第一个参数是文件名
		body = body + "sys.argv[" + index + "]"
		if i != argNum-1 {
			body += ","
		}
	}
	body = body + "))"

	body = appendCallBack(body, funcName+"-"+uuid)

	return []byte(body)
}

func appendRedirector(body string) string {
	body += "\nimport sys\n"
	body += "from urllib.parse import urlencode\n"
	body += "from urllib.request import urlopen\n"
	body += "with open('output.txt', 'w') as f:\n"
	body += "    sys.stdout = f\n"

	body += "    print(\"the result is:\")\n"
	return body
}

// 写都写了……
func appendCallBack(body string, name string) string {
	body += "\nurl='http://"
	body = body + _const.MASTER_IP + ":8080" + _const.SERVERLESS_CALLBACK_PATH + "'\n"
	body = body + "data = {'Name':'" + name + "'}\n"
	body += "s = urlencode(data)\n"
	body += "urlopen(url,s.encode())"
	return body
}

func addFunctionPod(pod *object.Pod) error {
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
