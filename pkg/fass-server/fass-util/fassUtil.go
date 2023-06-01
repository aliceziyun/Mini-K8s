package fass_util

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/third_party/file"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

func GetFunctionBody(body string, argNum int, serveName string, funcName string, uuid string) []byte {
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
	body = body + "))\n"

	body = appendCallBack(body, serveName+"_"+uuid)

	return []byte(body)
}

func appendRedirector(body string) string {
	body += "\nimport sys\n"
	body += "import requests\n"
	body += "with open('output.txt', 'w') as f:\n"
	body += "    sys.stdout = f\n"

	body += "    print(\"the result is:\")\n"
	return body
}

// 写都写了……
func appendCallBack(body string, name string) string {
	url := "http://"
	//这里不知道为什么只有内网才有回复
	url = url + _const.MATSTER_INNER_IP + ":8080" + _const.SERVERLESS_CALLBACK_PATH
	requestBody := fmt.Sprintf("requests.post(url='%s',headers={'Content-Type': 'application/x-www-form-urlencoded'},data={'name':'%s'})",
		url, name)
	body = body + requestBody
	return body
}

func AddFunctionPod(pod *object.Pod) error {
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

func InvokePod(uuid string, fileName string, dirName string, argList []string) error {
	fmt.Println(fileName)
	pod := object.Pod{}
	pod.Metadata.Name = fmt.Sprintf("Func-%s", uuid)
	pod.Metadata.Uid = uuid
	pod.Name = pod.Metadata.Name
	pod.Kind = object.POD

	container := object.Container{}
	container.Image = "testpy:latest"
	container.Name = fmt.Sprintf("Func-%s", uuid)
	//commands := []string{
	//	"/bin/sh",
	//	"-c",
	//	"while true; do echo hello world; sleep 1; done",
	//}
	commands := []string{
		"sudo",
		"python",
		fileName,
	}
	commands = append(commands, argList...)
	container.Command = commands
	//container.Args = commands
	volumeMounts := []object.VolumeMount{
		{
			Name:      "Serveless",
			MountPath: "/home/test",
		},
	}
	container.Ports = []object.ContainerPort{
		{Port: "6666"},
	}
	container.VolumeMounts = volumeMounts
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	volumes := []object.Volume{
		{
			Name: "Serveless",
			Type: "hostPath",
			Path: dirName,
		},
	}
	pod.Spec.Volumes = volumes

	go func() {
		fmt.Println("[FassServer] new function added!")
		err := AddFunctionPod(&pod)
		if err != nil {
			fmt.Println("[FassServer]", err)
			return
		}

	}()

	return nil
}

func InvokeFunction(meta *object.FunctionMeta, ls *listwatcher.ListWatcher) error {
	//从etcd中读取出Function实体
	nameAndUid := strings.Split(meta.Name, "_")
	name := nameAndUid[0]
	uuid := nameAndUid[1]

	resList, err := ls.List(_const.FUNC_CONFIG_PREFIX + "/" + name)

	if err != nil {
		return err
	}

	if len(resList) == 0 {
		return nil
	}
	function := &object.Function{}
	err = json.Unmarshal(resList[0].ValueBytes, function)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//读取函数
	zip, err := os.ReadFile(function.Path)
	functionBody := string(zip)

	body := GetFunctionBody(functionBody, function.ArgNum, function.Name, function.FuncName, uuid)

	//将文件放入共享目录中
	var fileName string
	var dirName string
	switch meta.Type {
	case "JOB":
		fileName = name + "_" + uuid + ".py"
		dirName = path.Join(_const.SHARED_DATA_DIR, "job-"+uuid)
		err = file.Bytes2File(body, fileName, dirName)
		break
	default:
		fileName = name + "_" + uuid + ".py"
		dirName = path.Join(_const.SHARED_DATA_DIR, name+"_"+uuid)
		err = file.Bytes2File(body, fileName, dirName)
		break
	}

	//创建Pod
	err = InvokePod(uuid, fileName, dirName, meta.ArgList)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
