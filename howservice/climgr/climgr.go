package climgr

import (
	"encoding/json"
	"sync"

	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

var (
	cliMap  map[string]*mode.Clipoint
	cliSync sync.Mutex
)

func init() {
	cliMap = make(map[string]*mode.Clipoint, 0)
}

type InnnerCall struct {
	mode.CallClimgr
}

func (c *InnnerCall) GetCliList() []string {
	ret := []string{}
	for k, _ := range cliMap {
		ret = append(ret, k)
	}
	return ret
}

func (c *InnnerCall) SendDataToDev(devId string, data interface{}) {

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
	for k, cli := range cliMap {
		if cli.Nid == nid {
			delete(cliMap, k)
			SyncOfflineToManager(cli.Name)
			log.Debugf("logout client:%s", cli.Name)
		}
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
	case "trans_data":
		transmsg(mapData, conn)
	}
}
