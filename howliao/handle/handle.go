package handle

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cocobao/howfw/howliao/conf"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

var (
	gconn *netconn.ClientConn
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
	case "list":
		list()
		return
	}

	if strings.HasPrefix(cmd, "@") && len(cmd) > 1 {
		if len(command) <= 1 {
			return
		}
		act, err := strconv.Atoi(cmd[1:])
		if err != nil {
			return
		}

		var msg string
		for _, v := range command[1:] {
			msg += v
			msg += " "
		}
		actmsg(act, msg)
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

func actmsg(act int, s string) {
	Send(map[string]interface{}{
		"cmd": "msg",
		"id":  act,
		"msg": s,
	})
}

func login(s []string) {
	serviceAddr := ApplyService(conf.GCfg.ManagerHost)
	if len(serviceAddr) > 0 {
		connectService(serviceAddr)
		Send(map[string]interface{}{
			"cmd":      "login",
			"username": s[0],
		})
	}
}

func list() {
	Send(map[string]interface{}{
		"cmd": "list",
	})
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
	defer fmt.Print("\n>>>")

	switch cmd {
	case "login":
		fmt.Println("login ok")
	case "list":
		if result, ok := mapData["result"].(map[string]interface{}); ok {
			for k, v := range result {
				fmt.Printf("\n%s: %v", k, v)
			}
			fmt.Print("\n")
		}
	case "msg":
		if result, ok := mapData["result"].(map[string]interface{}); ok {
			fmt.Print("\n")
			if from, ok := result["from"].(float64); ok {
				fmt.Printf("%v: ", from)
			}

			if msg, ok := result["msg"].(string); ok {
				fmt.Printf("%s\n", msg)
			}
		}
	}
}

func HeartBeat(time.Time, interface{}) {

}
