package service

import "github.com/cocobao/log"

func ApplyServiceHost() string {
	return rpcCli.ApplyServiceHost()
}

func devOnline(host string, md map[string]interface{}) {
	if v, ok := md["devid"].(string); ok {
		rpcCli.AddDev(host, v)
	}
	return
}

func devOffline(host string, md map[string]interface{}) {
	if v, ok := md["devid"].(string); ok {
		rpcCli.DecDev(host, v)
	}
}

func devTransData(host string, md map[string]interface{}) {
	var to_id string
	if v, ok := md["to_id"].(string); ok {
		to_id = v
	} else {
		log.Debug("no to id found")
		return
	}
	if len(to_id) == 0 {
		return
	}

	var from_id string
	if v, ok := md["from_id"].(string); ok {
		from_id = v
	}
	if len(from_id) == 0 {
		return
	}

	var data interface{}
	if v, ok := md["data"]; ok {
		data = v
	} else {
		log.Debug("no trans data found")
		return
	}

	serviceHost := rpcCli.DevBelongTo(to_id)
	if len(serviceHost) == 0 {
		log.Warnf("no serice for devid:%s found", to_id)
		return
	}

	log.Debugf("from:%s, to:%s, msg:%v, service:%s", from_id, to_id, data, serviceHost)

	callService(serviceHost, map[string]interface{}{
		"cmd":     "trans_data",
		"from_id": from_id,
		"to_id":   to_id,
		"data":    data,
	})
}
