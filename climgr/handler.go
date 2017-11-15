package climgr

import (
	"encoding/json"
	"strconv"

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
	if v, ok := mapData["username"].(string); ok {
		cliSync.Lock()
		cliMap[nid] = &mode.Clipoint{
			Name: v,
			Addr: c.RemoteAddr(),
		}
		cliSync.Unlock()
	}
}

func list(mapData map[string]interface{}, conn netconn.WriteCloser) {
	ls := make(map[string]interface{}, 0)
	cliSync.Lock()
	for k, v := range cliMap {
		id := strconv.Itoa(int(k))
		ls[id] = v.Name
	}
	cliSync.Unlock()
	Send(ls, conn)
}

func msg(mapData map[string]interface{}, conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	if _, ok := cliMap[nid]; !ok {
		log.Warnf("nid %v no login", nid)
		return
	}

	var id int64
	if v, ok := mapData["id"].(float64); ok {
		id = int64(v)
	} else {
		log.Warn("msg no id found", mapData)
		return
	}

	var msg string
	if v, ok := mapData["msg"].(string); ok {
		msg = v
	} else {
		log.Warn("msg no msg found", mapData)
		return
	}

	if v, ok := cliMap[id]; ok {
		Send(map[string]interface{}{
			"msg": msg,
		}, v.Conn)
	} else {
		log.Warn("peer client no login, nid:", id)
	}
}
