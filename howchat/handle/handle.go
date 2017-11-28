package handle

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cocobao/howfw/howchat/conf"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

var (
	gconn *netconn.ClientConn
	devId string
	count int64
	lastT int64
	speed int64
)

func SetCon(c *netconn.ClientConn) {
	gconn = c
}

func CommandHandle(command []string) {
	cmd := command[0]
	switch cmd {
	case "login":
		if len(command) <= 1 {
			return
		}
		login(command[1:])
		return
	}

	if strings.HasPrefix(cmd, "@") && len(cmd) > 1 {
		if len(command) <= 1 {
			return
		}

		var msg string
		for _, v := range command[1:] {
			msg += v
			msg += " "
		}
		actmsg(cmd[1:], msg)
	}
}

func Send(md map[string]interface{}) {
	if gconn == nil {
		log.Warn("conn not connect")
		return
	}
	data, err := json.Marshal(md)
	if err != nil {
		log.Warn("marshal md fail", err)
		return
	}
	gconn.Write(data)
}

func actmsg(toid string, s string) {
	var lt int64
	var sp int64
	var cn int64
	log.Debugf("act msg:%s to:%s", s, toid)
	for index := 0; index < 1000000; index++ {
		t := time.Now().Unix()
		if t > lt {
			lt = t
			sp = cn
			cn = 0
			fmt.Printf("speed:%d/s\n", sp)
		}
		cn++
		Send(map[string]interface{}{
			"cmd":     "trans_data",
			"from_id": devId,
			"to_id":   toid,
			"data":    s,
		})
		time.Sleep(100 * time.Microsecond)
	}
}

func login(s []string) {
	serviceAddr := ApplyService(conf.GCfg.ManagerHost)
	if len(serviceAddr) > 0 {
		connectService(serviceAddr)
		devId = s[0]
		Send(map[string]interface{}{
			"cmd":      "login",
			"username": devId,
		})
	}
}

func OnMessage(msg []byte, c netconn.WriteCloser) {
	var mapData map[string]interface{}
	if err := json.Unmarshal(msg, &mapData); err != nil {
		fmt.Println("unmarshal msg fail,", err)
		return
	}

	var cmd string
	if v, ok := mapData["cmd"].(string); ok {
		cmd = v
	} else {
		fmt.Println("result no cmd found")
		return
	}

	var status int
	if v, ok := mapData["status"].(float64); ok {
		status = int(v)
	} else {
		fmt.Println("result no status found")
		return
	}

	if status != 200 {
		fmt.Printf("\nresult cmd:%s status:%d\n", cmd, status)
		return
	}

	switch cmd {
	case "login":
		fmt.Println("login ok")
	case "trans_data":
		if result, ok := mapData["result"].(map[string]interface{}); ok {

			count++
			t := time.Now().Unix()
			if t > lastT {
				lastT = t
				speed = count
				count = 0
				fmt.Print("\n")
				if from, ok := result["from_id"].(string); ok {
					fmt.Printf("%v %d/s %v: ", t, speed, from)
				}
				if msg, ok := result["data"].(string); ok {
					fmt.Printf("%s\n", msg)
				}
			}
		}
	}
}

func HeartBeat(time.Time, interface{}) {

}
