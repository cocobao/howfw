package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cocobao/howfw/howeb/rpc"
	"github.com/cocobao/howfw/howeb/storage"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/log"
	"github.com/gin-gonic/gin"
)

func GetHome(g *gin.Context) {
	resp := &mode.TransResp{}
	rpc.CallManager(map[string]interface{}{"cmd": "dev_list"}, resp)

	var devList map[string][]string
	err := json.Unmarshal(resp.RespData.Body, &devList)
	log.Debugf("%d,%+v", resp.Code, devList)

	if err != nil || len(devList) == 0 {
		g.HTML(http.StatusOK, "home.html", gin.H{})
		return
	}

	var ret []interface{}
	var index int
	for k, v := range devList {
		servs := strings.Split(k, "/")
		for _, did := range v {
			data, err := storage.GetNewlestDevLog(did)
			if err == nil {
				index++
				data["service"] = servs[len(servs)-1]
				data["index"] = index
				ret = append(ret, data)
			} else {
				log.Warn(err)
			}
		}
	}
	log.Debugf("log:%+v", ret)
	g.HTML(http.StatusOK, "home.html", gin.H{"devlogs": ret})
}

func GetDevLog(g *gin.Context) {

}
