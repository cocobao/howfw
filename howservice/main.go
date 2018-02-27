package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/cocobao/howfw/howservice/climgr"
	"github.com/cocobao/howfw/howservice/conf"
	"github.com/cocobao/howfw/howservice/service"
	"github.com/cocobao/howfw/howservice/storage"
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/howfw/rpcser"
	"github.com/cocobao/howfw/signal"
	"github.com/cocobao/howfw/util/timeutil"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	storage.SetupMongoDB(conf.GCfg.MongoHost)

	rpcser.SetupEtcdConfig(
		conf.GCfg.LocalHost,
		conf.GCfg.EtcdServer.Username,
		conf.GCfg.EtcdServer.Password,
		"howservice",
		conf.GCfg.EtcdServer.Endpoints,
		conf.GCfg.EtcdServer.DialTimeout)
	rpcser.SetupRpcx(new(service.Trans), conf.GCfg.ServiceHost)
	go rpcser.RunRpcServer()
	go signal.GracefullyStopSever(func() {
		fmt.Println("~stop~~", timeutil.TimeToZoneStr(time.Now().Unix()))
		rpcser.StopRpcServer()
	})

	service.RunRpc(new(climgr.InnnerCall))
}

func main() {
	var tlsCfg *tls.Config
	if len(conf.GCfg.CerPath) > 0 {
		cer, err := tls.LoadX509KeyPair(conf.GCfg.CerPath+"/server.crt", conf.GCfg.CerPath+"/server.key")
		if err != nil {
			log.Println(err)
			return
		}
		tlsCfg = &tls.Config{Certificates: []tls.Certificate{cer}}
	}

	climgr.SetCallService(new(service.InnerCall))
	ser := netconn.NewServer(
		netconn.OnConnectOption(climgr.OnConnect),
		netconn.OnCloseOption(climgr.OnClose),
		netconn.OnMessageOption(climgr.OnMessage),
		netconn.TLSCredsOption(tlsCfg))
	s := strings.Split(conf.GCfg.ServiceHost, ":")
	if len(s) > 0 {
		ser.Start(s[1])
	} else {
		panic(s)
	}
}
