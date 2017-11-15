package main

import (
	"runtime"

	"github.com/cocobao/howfw/climgr"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.NewLogger("", log.LoggerLevelDebug)

	ser := netconn.NewServer(
		netconn.OnConnectOption(climgr.OnConnect),
		netconn.OnCloseOption(climgr.OnClose),
		netconn.OnMessageOption(climgr.OnMessage))
	ser.Start(50099)
}
