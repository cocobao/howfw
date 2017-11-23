package climgr

import (
	"encoding/json"
	"sync"

	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

var (
	cliMap  map[int64]*mode.Clipoint
	cliSync sync.Mutex
)

func init() {
	cliMap = make(map[int64]*mode.Clipoint, 0)
}

func OnConnect(conn netconn.WriteCloser) bool {
	return true
}

func OnClose(conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	log.Infof("client was close:%s, %d, %s", c.Name(), nid, c.RemoteAddr())

	cliSync.Lock()
	defer cliSync.Unlock()
	if _, ok := cliMap[nid]; ok {
		delete(cliMap, nid)
	}
}

func OnMessage(data []byte, conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		log.Error("unmarshal fail,", err)
		return
	}

	log.Debugf("nid:%d on msg:%v", nid, mapData)

	var cmd string
	if v, ok := mapData["cmd"].(string); ok {
		cmd = v
	}

	switch cmd {
	case "login":
		login(mapData, conn)
	case "list":
		list(mapData, conn)
	case "msg":
		msg(mapData, conn)
	}
}
