package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/cocobao/howfw/howmanager/climgr"
	"github.com/cocobao/howfw/howmanager/conf"
	"github.com/cocobao/howfw/howmanager/service"
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
		"howmanager",
		conf.GCfg.EtcdServer.Endpoints,
		conf.GCfg.EtcdServer.DialTimeout)
	rpcser.SetupRpcx(new(service.Trans), conf.GCfg.ServiceHost)
	go rpcser.RunRpcServer()
	go signal.GracefullyStopSever(func() {
		fmt.Println("~stop~~", timeutil.TimeToZoneStr(time.Now().Unix()))
		rpcser.StopRpcServer()
	})
	service.RunRpc()

	cliser := netconn.NewServer(
		netconn.OnConnectOption(climgr.OnConnect),
		netconn.OnCloseOption(climgr.OnClose),
		netconn.OnMessageOption(climgr.OnMessage))
	s := strings.Split(conf.GCfg.ServiceHost, ":")
	cliser.Start(s[1])
}
