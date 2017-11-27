package service

import "github.com/cocobao/log"

func devTransData(md map[string]interface{}) {
	var to_id string
	if v, ok := md["to_id"].(string); ok {
		to_id = v
	}
	if len(to_id) == 0 {
		return
	}

	var data string
	if v, ok := md["data"].(string); ok {
		data = v
	} else {
		return
	}

	log.Debug("to_id:%s, data:%v", to_id, data)

	// mid.SendDataToDev(to_id, data)
}
