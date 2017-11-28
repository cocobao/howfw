package service

import "github.com/cocobao/log"

func devTransData(md map[string]interface{}) {
	var to_id string
	if v, ok := md["to_id"].(string); ok {
		to_id = v
	}
	if len(to_id) == 0 {
		log.Warn("no to id found")
		return
	}

	var data map[string]interface{}
	if v, ok := md["data"].(map[string]interface{}); ok {
		data = v
	} else {
		log.Warn("no data found")
		return
	}

	log.Debugf("to_id:%s, data:%v", to_id, data)

	callClimgr.SendDataToDev(to_id, data)
}
