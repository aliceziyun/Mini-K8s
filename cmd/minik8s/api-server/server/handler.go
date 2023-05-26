package apiserver

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func (s *APIServer) watch(ctx *gin.Context) {
	//TODO: 现在是超简化版本
	key := ctx.Request.URL.Path
	fmt.Printf("[API-Server] receive watch request with key %s \n", key)
	s.watcherChan <- watchOpt{key: key, withPrefix: false}
	ctx.Data(http.StatusOK, "application/json", nil)
}

func (s *APIServer) put(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	err = s.store.Put(key, string(body))
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *APIServer) delete(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	err := s.store.Del(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *APIServer) addPodConfig(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := &object.Pod{}
	err = json.Unmarshal(body, pod)
	//这里为了方便replicaSet等设置成running，实际有没有挂掉靠监听同步
	pod.Status.Phase = object.RUNNING
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	key := _const.POD_CONFIG_PREFIX + "/" + pod.Name
	fmt.Printf("key:%v\n", key)

	body, _ = json.Marshal(pod)

	err = s.store.Put(key, string(body))
	if err != nil {
		return
	}
}

func (s *APIServer) deletePod(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	var name string
	err = json.Unmarshal(body, &name)

	key := _const.POD_CONFIG_PREFIX + "/" + name
	resList, err := s.store.Get(key)
	if err != nil || len(resList) == 0 {
		fmt.Printf("[API-Server] pod not exist:%s\n", name)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	pod := object.Pod{}
	err = json.Unmarshal(resList[0].ValueBytes, &pod)
	if err != nil {
		fmt.Printf("[API-Server] pod unmarshal fail\n")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	pod.Status.Phase = object.DELETED
	raw, _ := json.Marshal(pod)
	fmt.Println("[API-Server] delete pod ", pod.Name)
	err = s.store.Put(key, string(raw))
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *APIServer) addPodRuntime(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := &object.Pod{}
	err = json.Unmarshal(body, pod)
	pod.Status.Phase = object.RUNNING
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	key := _const.POD_RUNTIME_PREFIX + "/" + pod.Name
	fmt.Printf("key:%v\n", key)

	body, _ = json.Marshal(pod)

	err = s.store.Put(key, string(body))
	if err != nil {
		return
	}
}

func (s *APIServer) addRS(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	rs := &object.ReplicaSet{}
	err = json.Unmarshal(body, rs)
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	rs.Status.Status = object.RUNNING

	key := _const.RS_CONFIG_PREFIX + "/" + rs.Name
	fmt.Printf("key:%v\n", key)

	body, _ = json.Marshal(rs)

	err = s.store.Put(key, string(body))
	if err != nil {
		return
	}
}

func (s *APIServer) deleteRS(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	var name string
	err = json.Unmarshal(body, &name)

	key := _const.RS_CONFIG_PREFIX + "/" + name
	resList, err := s.store.Get(key)
	if err != nil || len(resList) == 0 {
		fmt.Printf("[API-Server] RS not exist:%s\n", name)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rs := object.ReplicaSet{}
	err = json.Unmarshal(resList[0].ValueBytes, &rs)
	if err != nil {
		fmt.Printf("[API-Server] unmarshal fail\n")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rs.Status.Status = object.DELETED
	raw, _ := json.Marshal(rs)
	fmt.Println("[API-Server] delete rs ", rs.Name)
	err = s.store.Put(key, string(raw))
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *APIServer) get(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	listRes, err := s.store.Get(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	data, err := json.Marshal(listRes)
	ctx.Data(http.StatusOK, "application/json", data)
}

func (s *APIServer) getByPrefix(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	listRes, err := s.store.GetPrefix(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	data, err := json.Marshal(listRes)
	ctx.Data(http.StatusOK, "application/json", data)
}
