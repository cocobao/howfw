package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cocobao/howfw/howeb/conf"
	"github.com/cocobao/howfw/howeb/router"
	"github.com/cocobao/howfw/howeb/rpc"
	"github.com/cocobao/howfw/howeb/storage"
	"github.com/cocobao/howfw/rpcser"
	"github.com/facebookgo/grace/gracehttp"
)

func main() {
	storage.SetupMongoDB(conf.GCfg.MongoHost)

	rpcser.SetupEtcdConfig(
		"howeb"+conf.GCfg.LocalPort,
		conf.GCfg.EtcdServer.Username,
		conf.GCfg.EtcdServer.Password,
		"howeb",
		conf.GCfg.EtcdServer.Endpoints,
		conf.GCfg.EtcdServer.DialTimeout)
	rpc.RunRpc()

	err := gracehttp.Serve(
		&http.Server{
			Addr:    conf.GCfg.LocalPort,
			Handler: router.LoadRouter(),
		},
	)
	if err != nil {
		fmt.Println(err, "setup api service fail")
		os.Exit(0)
	}
}
