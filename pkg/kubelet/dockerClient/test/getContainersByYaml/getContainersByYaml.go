// package main

package parseYaml

import (
	"Mini-K8s/pkg/object"
	"fmt"
	"io/ioutil"

	v2 "gopkg.in/yaml.v2"
)

var (
	cfgFile = "config.yaml"
)

type user struct {
	Id       string   `yaml:"id"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`
}

type Config struct {
	Listen    string            `yaml:"listen"`
	SecretKey int               `yaml:"secret_key"`
	Boll      bool              `yaml:"bowls"`
	StrSlice  []string          `yaml:"strslice"`
	Auth      []user            `yaml:"auth"`
	KeyMap    map[string]string `yaml:"keymap"`
}

type PodConfig struct {
	Kind string            `yaml:"kind"`
	Spec object.Containers `yaml:"spec"`
	// Containers []object.Container `yaml:"containers"`
}

func GetContainersByFile(path string) []object.Container {
	return getContainersByFile(path)
}

func getContainersByFile(path string) []object.Container {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	podCfg := new(PodConfig)
	err = v2.Unmarshal([]byte(data), podCfg)
	if err != nil {
		fmt.Printf("file in %s unmarshal fail, use default config", path)
		return nil
	}
	return podCfg.Spec.Containers
}

func test() {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	conf := new(Config)
	if err := v2.Unmarshal(data, conf); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("conf: %v\n", conf)
	fmt.Printf("conf.SecretKey: %v\n", conf.SecretKey)

	out, err := v2.Marshal(conf)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("out: %v\n", string(out))
	return
}

func main() {
	// test()
	path := "testPod.yaml"
	// path := "../../build/pod/testPod.yaml"

	containers := getContainersByFile(path)
	for _, value := range containers {
		fmt.Printf(
			"========Container Info:======== \nname=%s\nimage=%s\n",
			value.Image,
			value.Name,
		)
	}
}
