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

func OnConnect(conn netconn.WriteCloser) bool {
	return true
}

func OnClose(conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	log.Infof("client was close:%s, %d, %s", c.Name(), nid, c.RemoteAddr())

	cliSync.Lock()
	if _, ok := cliMap[nid]; ok {
		delete(cliMap, nid)
	}
	cliSync.Unlock()
}

func OnMessage(data []byte, conn netconn.WriteCloser) {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		log.Error("unmarshal fail,", err)
		return
	}

	var cmd string
	if v, ok := mapData["cmd"].(string); ok {
		cmd = v
	}

	switch cmd {
	case "login":
		login(mapData, conn)
	}
}
