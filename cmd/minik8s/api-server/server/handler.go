package apiserver

import (
	"Mini-K8s/pkg/object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (s *APIServer) watch(ctx *gin.Context) {
	//key := ctx.Request.URL.Path
	ticketStr, status := ctx.GetPostForm("ticket")
	fmt.Println(ticketStr, status)
	if !status {
		t := Ticket{}
		t.T = s.ticketSeller.Add(1)
		data, _ := json.Marshal(t)
		//s.watcherChan <- watchOpt{key: key, withPrefix: false, ticket: t.T}
		ctx.Data(http.StatusOK, "application/json", data)
	} else {
		s.watcherMtx.Lock()
		ticket, err := strconv.ParseUint(ticketStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			ctx.AbortWithStatus(http.StatusBadRequest)
		} else {
			//if s.watcherMap[key] != nil {
			//	s.watcherMap[key].set.Remove(ticket)
			//	if s.watcherMap[key].set.Equal(mapset.NewSet[uint64]()) {
			//		s.watcherMap[key].cancel()
			//		s.watcherMap[key] = nil
			//		klog.Infof("Cancel the watcher of key %s\n", key)
			//	}
			//}
			fmt.Println(ticket, "ok")
			ctx.Status(http.StatusOK)
		}
		s.watcherMtx.Unlock()
	}
}

func (s *APIServer) addPodTest(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := &object.Pod{}
	err = json.Unmarshal(body, pod)
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	key := "test"
	fmt.Printf("key:%v\n", key)

	body, _ = json.Marshal(pod)

	err = s.store.Put(key, string(body))
	if err != nil {
		return
	}
}
