package climgr

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cocobao/howfw/howmanager/service"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/howfw/util/timer"
	"github.com/cocobao/log"
)

var (
	cmg *ClientMgr
	one *sync.Once
)

func init() {
	one = &sync.Once{}
}

//获取客户端管理器单例
func GetClientMgr() *ClientMgr {
	one.Do(func() {
		cmg = &ClientMgr{
			clis:   make([]*mode.Clipoint, 0),
			freCtl: make(map[string]int, 0),
		}

		timer.GetTimingWheel().AddTimer(time.Now(), 5*time.Second, &timer.OnTimeOut{
			Ctx:      context.Background(),
			Callback: cmg.OnTimer,
		})
	})
	return cmg
}

type ClientMgr struct {
	//客户端节点列表
	clis    []*mode.Clipoint
	cliSync sync.Mutex

	//频率控制
	freCtl map[string]int
}

//定时器检测
func (c *ClientMgr) OnTimer(time.Time, interface{}) {
	// fmt.Println(time.Now().Unix())

	//清除频控表
	c.cliSync.Lock()
	if len(c.freCtl) > 0 {
		c.freCtl = map[string]int{}
	}

	//每个ip连接不能超过5秒
	now := time.Now().Unix()
	var needClose []*mode.Clipoint
	if len(c.clis) > 0 {
		for index := len(c.clis) - 1; index > 0; index-- {
			cli := c.clis[index]
			if now-cli.ConTime > 5 {
				log.Warnf("client %s link more than 5 sec")
				needClose = append(needClose, cli)
				c.clis = append(c.clis[:index], c.clis[index+1:]...)
			}
		}
	}
	c.cliSync.Unlock()

	for _, v := range needClose {
		v.Conn.Close()
	}
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
		Addr:    addr,
		Conn:    conn,
		Nid:     nid,
		ConTime: time.Now().Unix(),
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

	//删除频控
	if _, ok := c.freCtl[addr]; ok {
		delete(c.freCtl, addr)
	}

	index := -1
	for i, v := range c.clis {
		if v.Addr == addr {
			index = i
			break
		}
	}
	//删除客户端
	if index > 0 {
		c.clis = append(c.clis[:index], c.clis[index+1:]...)
	}
}

//接收数据
func (c *ClientMgr) OnMessage(data []byte, conn netconn.WriteCloser) {
	addr := conn.(*netconn.ServerConn).RemoteAddr()

	c.cliSync.Lock()
	//频率控制
	if v, ok := c.freCtl[addr]; ok {
		v++
		c.freCtl[addr] = v
		//5秒内请求不超过一定次数
		if v > 100 {
			c.cliSync.Unlock()
			log.Warn("client :%s apply too mach", addr)
			conn.Close()
			return
		}
	}
	c.cliSync.Unlock()

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

//请求登录
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
