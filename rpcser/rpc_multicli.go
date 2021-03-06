package rpcser

import (
	"fmt"
	"sync"
	"time"

	"github.com/cocobao/log"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
)

type RpcxMultiCli struct {
	ServiceName string
	Client      *rpcx.Client
	SyLock      sync.Mutex
}

func (r *RpcxMultiCli) GetMultiClient() *rpcx.Client {
	r.SyLock.Lock()
	defer r.SyLock.Unlock()

	if r.Client == nil {
		r.Client = newMultiRpcxClient(r.ServiceName)
		if r.Client != nil {
			EtcdWatch(r.ServiceName, false, func(t int, k string, v map[string]interface{}) {
				r.SyLock.Lock()
				defer r.SyLock.Unlock()
				log.Warn("client release,", k, v)
				r.Client = nil
			})
		}
	}
	return r.Client
}

//new一个rpcx客户端
func newMultiRpcxClient(sname string) *rpcx.Client {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	etcdConfs := GetEtcdServiceList(sname)
	if etcdConfs == nil || len(etcdConfs) == 0 {
		return nil
	}

	servers := []*clientselector.ServerPeer{}
	for url, _ := range etcdConfs {
		if len(url) <= 10 {
			continue
		}
		servers = append(servers, &clientselector.ServerPeer{Network: "tcp", Address: url})
	}

	if len(servers) == 0 {
		log.Warn("no servers found, %+v", etcdConfs)
		return nil
	}

	client := rpcx.NewClient(clientselector.NewMultiClientSelector(servers, rpcx.RandomSelect, 10*time.Second))
	if client == nil {
		log.Warn("rpcx.NewClient fail")
		return nil
	}
	log.Debug("new rpcx ok:", etcdConfs)
	client.FailMode = rpcx.Failover
	return client
}
