package service

import (
	"encoding/json"

	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/log"
)

type Trans int

func (t *Trans) TransIn(req *mode.TransData, reply *mode.TransResp) (err error) {
	defer func() {
		reply.Code = 200
		reply.Err = ""
	}()
	var val map[string]interface{}
	if e := json.Unmarshal(req.Body, &val); e != nil {
		log.Warn("")
		return
	}

	var cmd string
	if v, ok := val["cmd"].(string); ok {
		cmd = v
	}
	// log.Debugf("body:%+v", val)

	var host string
	if v, ok := req.Headers["host"]; ok {
		host = v
	}

	switch cmd {
	//设备上线通知
	case "dev_online":
		devOnline(host, val)
	//设备离线通知
	case "dev_offline":
		devOffline(host, val)
	//消息透传
	case "trans_data":
		devTransData(host, val)
	//设备列表上报
	case "dev_list":
		l := devList()
		log.Debug("dev list:", l)
		data, _ := json.Marshal(l)
		reply.RespData = mode.TransData{
			Body: data,
		}
	default:
		log.Debug("invalid cmd:", cmd)
	}
	return
}
