package service

import (
	"encoding/json"

	"github.com/cocobao/howfw/howmanager/conf"
	"github.com/cocobao/howfw/mode"
	"github.com/cocobao/howfw/rpcser"
	"github.com/cocobao/log"
)

var (
	rpcCli *rpcser.RpcxClis
)

func RunRpc() {
	rpcCli = &rpcser.RpcxClis{
		PrefixName:    "/howservice",
		OnConnectCall: onConnect,
	}
	rpcCli.LoadClients()
}

func callService(host string, md map[string]interface{}) {
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
	rpcCli.SendDataToClient(host, "howservice.TransIn", data, &mode.TransResp{})
}

//服务连上之后，进行客户端列表同步
func onConnect(cliKey string) {
	data := mode.TransData{
		Headers: map[string]string{
			"host": conf.GCfg.LocalHost,
		},
		Body: nil,
	}
	var reply mode.TransResp
	rpcCli.CallClient(cliKey, "howservice.SynDevlist", data, &reply)

	var body map[string]interface{}
	if err := json.Unmarshal(reply.RespData.Body, &body); err != nil {
		log.Warn(err)
		return
	}
	log.Debugf("body:%v", body)

	if v, ok := body["devlist"].([]interface{}); ok && len(v) > 0 {
		for _, d := range v {
			rpcCli.AddDevWithKey(cliKey, d.(string))
		}
	}
}
