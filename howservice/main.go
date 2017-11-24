package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/cocobao/howfw/howservice/climgr"
	"github.com/cocobao/howfw/howservice/conf"
	"github.com/cocobao/howfw/howservice/rpcservice"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/howfw/rpcser"
	"github.com/cocobao/howfw/signal"
	"github.com/cocobao/howfw/util/timeutil"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rpcser.SetupEtcdConfig(
		conf.GCfg.LocalHost,
		conf.GCfg.EtcdServer.Username,
		conf.GCfg.EtcdServer.Password,
		"howservice",
		conf.GCfg.EtcdServer.Endpoints,
		conf.GCfg.EtcdServer.DialTimeout)
	rpcser.SetupRpcx(new(rpcservice.Trans), conf.GCfg.ServiceHost)
	go rpcser.RunRpcServer()
	go signal.GracefullyStopSever(func() {
		fmt.Println("~stop~~", timeutil.TimeToZoneStr(time.Now().Unix()))
		rpcser.StopRpcServer()
	})
	rpcservice.RunRpc()

	ser := netconn.NewServer(
		netconn.OnConnectOption(climgr.OnConnect),
		netconn.OnCloseOption(climgr.OnClose),
		netconn.OnMessageOption(climgr.OnMessage))
	s := strings.Split(conf.GCfg.ServiceHost, ":")
	ser.Start(s[1])
}
