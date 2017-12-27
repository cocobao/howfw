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
	cmg *ClientMgr
	one *sync.Once
)

func init() {
	one = &sync.Once{}
}

func GetClientMgr() *ClientMgr {
	one.Do(func() {
		cmg = &ClientMgr{
			clis:   make([]*mode.Clipoint, 0),
			freCtl: make(map[string]int, 0),
		}
	})
	return cmg
}

type ClientMgr struct {
	clis    []*mode.Clipoint
	cliSync sync.Mutex
	freCtl  map[string]int
}

func (c *ClientMgr) Send(md map[string]interface{}, conn netconn.WriteCloser) {
	data, err := json.Marshal(md)
	if err != nil {
		log.Warn("marshal md fail", err)
		return
	}
	conn.Write(data)
}

func (c *ClientMgr) OnConnect(conn netconn.WriteCloser) bool {
	con := conn.(*netconn.ServerConn)
	nid := con.NetID()
	addr := con.RemoteAddr()

	c.cliSync.Lock()
	defer c.cliSync.Unlock()

	//是否已经存在，如果有，先删除
	index := -1
	for i, v := range c.clis {
		if v.Addr == addr {
			index = i
			break
		}
	}

	if index > 0 {
		c.clis = append(c.clis[:index], c.clis[index+1:]...)
	}

	c.clis = append(c.clis, &mode.Clipoint{
		Addr: addr,
		Conn: conn,
		Nid:  nid,
	})

	return true
}

func (c *ClientMgr) OnClose(conn netconn.WriteCloser) {
	con := conn.(*netconn.ServerConn)
	nid := con.NetID()
	addr := con.RemoteAddr()
	log.Infof("client was close:%s, %d, %s", con.Name(), nid, con.RemoteAddr())

	c.cliSync.Lock()
	defer c.cliSync.Unlock()

	index := -1
	for i, v := range c.clis {
		if v.Addr == addr {
			index = i
			break
		}
	}

	if index > 0 {
		c.clis = append(c.clis[:index], c.clis[index+1:]...)
	}
}

func (c *ClientMgr) OnMessage(data []byte, conn netconn.WriteCloser) {
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
	//请求服务节点
	case "apply":
		c.apply(mapData, conn)
	}
}

func (c *ClientMgr) apply(mapData map[string]interface{}, conn netconn.WriteCloser) {
	defer conn.Close()

	//获取一个服务节点
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
	c.Send(md, conn)
}
