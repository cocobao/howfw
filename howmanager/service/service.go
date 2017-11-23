package service

import "github.com/cocobao/howfw/rpcser"

var (
	rpcCli *rpcser.RpcxClis
)

func RunRpc() {
	rpcCli = &rpcser.RpcxClis{
		PrefixName: "/howservice",
	}
	rpcCli.GetClients()
}

func ApplyServiceHost() string {
	return rpcCli.ApplyServiceHost()
}
