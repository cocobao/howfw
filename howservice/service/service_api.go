package service

import (
	"encoding/json"

	"github.com/cocobao/howfw/howservice/conf"
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

	log.Debugf("call trans in:%v", val)

	var cmd string
	if v, ok := val["cmd"].(string); ok {
		cmd = v
	}

	switch cmd {
	case "trans_data":
		devTransData(val)
	}
	return
}

func (t *Trans) SynDevlist(req *mode.TransData, reply *mode.TransResp) (err error) {
	defer func() {
		reply.Code = 200
		reply.Err = ""
	}()
	devList := callClimgr.GetCliList()

	data, _ := json.Marshal(map[string]interface{}{
		"devlist": devList,
	})
	reply.RespData.Body = data
	reply.RespData.Headers = map[string]string{
		"host": conf.GCfg.LocalHost,
	}
	return
}

func (t *Trans) ReqFromWebservice(req *mode.TransData, reply *mode.TransResp) (err error) {
	defer func() {
		reply.Code = 200
		reply.Err = ""
	}()

}
