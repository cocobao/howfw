package climgr

import (
	"encoding/json"
	"time"

	"github.com/cocobao/howfw/howservice/storage"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

type handleData struct {
	data []byte
	conn netconn.WriteCloser
}

var (
	handleChan chan *handleData
)

func init() {
	handleChan = make(chan *handleData, 10000)
	taskHandle()
}

func taskHandle() {
	go func() {
		for {
			msg := <-handleChan

			// c := msg.conn.(*netconn.ServerConn)
			// nid := c.NetID()
			var mapData map[string]interface{}
			if err := json.Unmarshal(msg.data, &mapData); err != nil {
				log.Error("unmarshal fail,", err)
				return
			}

			// log.Debugf("nid:%d on msg:%v", nid, mapData)

			var cmd string
			if v, ok := mapData["cmd"].(string); ok {
				cmd = v
			}

			switch cmd {
			case "login":
				login(mapData, msg.conn)
			case "trans_data":
				transmsg(mapData, msg.conn)
			case "report_event":
				report_event(mapData, msg.conn)
			}
		}
	}()
}

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

	// log.Debugf("trans data from:%s to:%s data:%v", fromid, toid, msgData)

	if v, ok := cliMap[toid]; ok {
		//manager转发
		Send(msgData, v.Conn)
	} else {
		//内部转发
		TransMsgToDev(fromid, toid, msgData)
	}
}

func report_event(mapData map[string]interface{}, conn netconn.WriteCloser) {
	var devType string
	if v, ok := mapData["dev_type"].(string); ok {
		devType = v
	}

	var devId string
	if v, ok := mapData["dev_id"].(string); ok {
		devId = v
	} else {
		log.Warn("no dev_id found")
		return
	}

	storage.InsertEventData(map[string]interface{}{
		"device_id":   devId,
		"device_type": devType,
		"insert_time": time.Now().Format("2006-01-02T15:04:05-07:00"),
		"event":       mapData,
	})
}
