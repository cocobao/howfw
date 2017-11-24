package service

import "github.com/cocobao/howfw/rpcser"

var (
	rpcCli *rpcser.RpcxClis
)

func RunRpc() {
	rpcCli = &rpcser.RpcxClis{
		PrefixName: "/howservice",
	}
	rpcCli.LoadClients()
}

func ApplyServiceHost() string {
	return rpcCli.ApplyServiceHost()
}

func devOnline(host string, md map[string]interface{}) {
	if v, ok := md["devid"].(string); ok {
		rpcCli.AddDev(host, v)
	}
	return
}

func devOffline(host string, md map[string]interface{}) {
	if v, ok := md["devid"].(string); ok {
		rpcCli.DecDev(host, v)
	}
}
