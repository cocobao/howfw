package climgr

import (
	"encoding/json"
	"sync"

	"github.com/cocobao/howfw/howmanager/service"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

var (
	clis    []*mode.Clipoint
	cliSync sync.Mutex
	roopID  int
)

func init() {
	roopID = 0
	clis = make([]*mode.Clipoint, 0)
}

func getRoopID() int {
	roopID++
	return roopID
}

func Send(md map[string]interface{}, conn netconn.WriteCloser) {
	data, err := json.Marshal(md)
	if err != nil {
		log.Warn("marshal md fail", err)
		return
	}
	conn.Write(data)
}

func OnConnect(conn netconn.WriteCloser) bool {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	addr := c.RemoteAddr()

	cliSync.Lock()
	defer cliSync.Unlock()

	//是否已经存在，如果有，先删除
	index := -1
	for i, v := range clis {
		if v.Addr == addr {
			index = i
			break
		}
	}

	if index > 0 {
		clis = append(clis[:index], clis[index+1:]...)
	}

	clis = append(clis, &mode.Clipoint{
		Addr: addr,
		Conn: conn,
		Nid:  nid,
	})

	return true
}

func OnClose(conn netconn.WriteCloser) {
	c := conn.(*netconn.ServerConn)
	nid := c.NetID()
	addr := c.RemoteAddr()
	log.Infof("client was close:%s, %d, %s", c.Name(), nid, c.RemoteAddr())

	cliSync.Lock()
	defer cliSync.Unlock()

	index := -1
	for i, v := range clis {
		if v.Addr == addr {
			index = i
			break
		}
	}

	if index > 0 {
		clis = append(clis[:index], clis[index+1:]...)
	}
}

func OnMessage(data []byte, conn netconn.WriteCloser) {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		log.Error("unmarshal fail,", err)
		return
	}

	log.Debugf("nid:%d on msg:%v", conn.(*netconn.ServerConn).NetID(), mapData)

	var cmd string
	if v, ok := mapData["cmd"].(string); ok {
		cmd = v
	}

	switch cmd {
	case "apply":
		apply(mapData, conn)
	}
}

func apply(mapData map[string]interface{}, conn netconn.WriteCloser) {
	host := service.ApplyServiceHost()

	var md map[string]interface{}
	if len(host) <= 0 {
		md = map[string]interface{}{
			"cmd":    "apply",
			"status": 701,
			"result": "",
		}
	} else {
		md = map[string]interface{}{
			"cmd":    "apply",
			"status": 200,
			"result": map[string]interface{}{
				"saddr": host,
			},
		}
	}
	Send(md, conn)
}
