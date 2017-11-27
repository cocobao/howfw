package climgr

import (
	"encoding/json"

	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

func Send(md map[string]interface{}, conn netconn.WriteCloser) {
	data, err := json.Marshal(md)
	if err != nil {
		log.Warn("marshal md fail", err)
		return
	}
	conn.Write(data)
}

func login(mapData map[string]interface{}, conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	addr := c.RemoteAddr()

	cliSync.Lock()
	defer cliSync.Unlock()
	if v, ok := mapData["username"].(string); ok {
		log.Debugf("nid:%d, addr:%s, user:%s", nid, addr, v)

		cliMap[v] = &mode.Clipoint{
			Nid:   nid,
			Name:  v,
			Addr:  addr,
			Conn:  conn,
			Binds: []string{},
		}

		Send(map[string]interface{}{
			"cmd":    "login",
			"status": 200,
		}, conn)

		SyncOnlineToManager(v)
	}
}

func transmsg(mapData map[string]interface{}, conn netconn.WriteCloser) {
	var fromid string
	if v, ok := mapData["from_id"].(string); ok {
		fromid = v
	} else {
		log.Warn("msg no id found", mapData)
		return
	}

	var toid string
	if v, ok := mapData["to_id"].(string); ok {
		toid = v
	} else {
		log.Debug("no to id found")
		return
	}

	if len(fromid) == 0 || len(toid) == 0 {
		log.Debug("id invalid")
		return
	}

	var msg string
	if v, ok := mapData["data"].(string); ok {
		msg = v
	} else {
		log.Warn("msg no msg found", mapData)
		return
	}

	msgData := map[string]interface{}{
		"cmd":    "trans_data",
		"status": 200,
		"result": map[string]interface{}{
			"from_id": fromid,
			"to_id":   toid,
			"data":    msg,
		},
	}

	log.Debugf("trans data from:%s to:%s data:%v", fromid, toid, msgData)

	// if v, ok := cliMap[toid]; ok {
	// 	Send(msgData, v.Conn)
	// } else {
	TransMsgToDev(fromid, toid, msgData)
	// }
}
