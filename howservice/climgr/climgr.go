package climgr

import (
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
}

func (c *InnnerCall) GetCliList() []string {
	ret := []string{}
	for k, _ := range cliMap {
		ret = append(ret, k)
	}
	return ret
}

func (c *InnnerCall) SendDataToDev(devId string, data map[string]interface{}) {
	if v, ok := cliMap[devId]; ok {
		Send(data, v.Conn)
	} else {
		log.Warnf("no service for devid:%s found")
	}
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
	handleChan <- &handleData{
		data: data,
		conn: conn,
	}
}
