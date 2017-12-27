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
}

func (c *InnerCall) CallManager(isBroadcast bool, md map[string]interface{}) {
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

	if isBroadcast {
		allClis := cli.ClientSelector.AllClients(cli.ClientCodecFunc)
		for _, v := range allClis {
			if err := v.Call(context.Background(), "howmanager.TransIn", data, &mode.TransResp{}); err != nil {
				log.Warn("call transin fail", err)
				continue
			}
		}
	} else {
		if err := cli.Call(context.Background(), "howmanager.TransIn", data, &mode.TransResp{}); err != nil {
			log.Warn("call transin fail", err)
			return
		}
	}
}
