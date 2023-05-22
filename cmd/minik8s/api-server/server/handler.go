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

func (s *APIServer) addPod(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := &object.Pod{}
	err = json.Unmarshal(body, pod)
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

func (s *APIServer) addRS(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	rs := &object.ReplicaSet{}
	err = json.Unmarshal(body, rs)
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	key := _const.RS_CONFIG_PREFIX + "/" + rs.Name
	fmt.Printf("key:%v\n", key)

	body, _ = json.Marshal(rs)

	err = s.store.Put(key, string(body))
	if err != nil {
		return
	}
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
