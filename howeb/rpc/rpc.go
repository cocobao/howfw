package rpc

import (
	"context"
	"encoding/json"

	"github.com/cocobao/howfw/howeb/conf"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/rpcser"
	"github.com/cocobao/log"
)

var (
	rpcCli *rpcser.RpcxMultiCli
)

func RunRpc() {
	rpcCli = &rpcser.RpcxMultiCli{
		ServiceName: "/howmanager",
	}
}

func CallManager(md map[string]interface{}, resp *mode.TransResp) {
	tmpData, err := json.Marshal(md)
	if err != nil {
		log.Warn("marsha data fail", err)
		return
	}
	data := mode.TransData{
		Headers: map[string]string{
			"host": conf.GCfg.LocalPort,
		},
		Body: tmpData,
	}
	cli := rpcCli.GetMultiClient()
	if cli == nil {
		log.Warn("no manager cli found")
		return
	}
	if err := cli.Call(context.Background(), "howmanager.TransIn", data, resp); err != nil {
		log.Warn("call transin fail", err)
		return
	}
}
