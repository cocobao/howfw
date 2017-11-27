package service

import (
	"context"
	"encoding/json"

	"github.com/cocobao/howfw/howservice/conf"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/rpcser"
	"github.com/cocobao/log"
)

var (
	rpcCli     *rpcser.RpcxMultiCli
	callClimgr mode.CallClimgr
)

func RunRpc(c mode.CallClimgr) {
	rpcCli = &rpcser.RpcxMultiCli{
		ServiceName: "/howmanager",
	}
	callClimgr = c
}

type InnerCall struct {
	mode.CallService
}

func (c *InnerCall) CallManager(md map[string]interface{}) {
	tmpData, err := json.Marshal(md)
	if err != nil {
		log.Warn("marsha data fail", err)
		return
	}
	data := mode.TransData{
		Headers: map[string]string{
			"host": conf.GCfg.LocalHost,
		},
		Body: tmpData,
	}
	cli := rpcCli.GetMultiClient()
	if cli == nil {
		log.Warn("no manager cli found")
		return
	}
	if err := cli.Call(context.Background(), "howmanager.TransIn", data, &mode.TransResp{}); err != nil {
		log.Warn("call transin fail", err)
		return
	}
}
