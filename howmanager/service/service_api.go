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
	case "dev_online":
		devOnline(host, val)
	case "dev_offline":
		devOffline(host, val)
	case "trans_data":
		devTransData(host, val)
	case "dev_list":
		l := devList()
		log.Debug("dev list:", l)
		data, _ := json.Marshal(l)
		reply.RespData = mode.TransData{
			Body: data,
		}
	}
	return
}
