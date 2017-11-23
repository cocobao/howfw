package rpcser

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	"github.com/cocobao/log"
	"github.com/coreos/etcd/clientv3"
	"github.com/smallnest/rpcx"
)

var (
	etcdClient *clientv3.Client
)

type RpcxClis struct {
	PrefixName string
	Clients    map[string]*rpcx.Client
	MapData    map[string]map[string]interface{}
	devList    map[string][]string
}

func (r *RpcxClis) ApplyServiceHost() string {
	if r.Clients == nil || len(r.Clients) == 0 {
		return ""
	}
	less := math.MaxUint32
	kk := ""
	for k, v := range r.devList {
		l := len(v)
		if l < less {
			less = l
			kk = k
		}
	}

	if v, ok := r.MapData[kk]; ok {
		if sh, ok := v["service_host"].(string); ok {
			return sh
		}
	}

	return ""
}

func (r *RpcxClis) GetClients() {
	if r.Clients == nil {
		r.Clients = newRpcxClients(r.PrefixName)
		if r.Clients == nil {
			r.Clients = make(map[string]*rpcx.Client, 0)
		}
		r.MapData = make(map[string]map[string]interface{}, 0)
		r.devList = make(map[string][]string, 0)
		EtcdWatch(r.PrefixName, true, r.changeCalls)
	}
}

func (r *RpcxClis) changeCalls(t int, k string, v map[string]interface{}) {
	if t == 0 {
		c := newRpcxClient(k, r.changeCalls)
		r.Clients[k] = c
		r.MapData[k] = v
		r.devList[k] = []string{}
		log.Debug("new rpcx client,", k, v)
	} else if t == 1 {
		if vc, ok := r.Clients[k]; ok {
			vc.Close()
			delete(r.Clients, k)
			delete(r.MapData, k)
			delete(r.devList, k)
			log.Warn("delete rpcx cient,", k, v)
			return
		}
	}
}

func EtcdClient() *clientv3.Client {
	if etcdClient == nil {
		acfg := &AuthCfg{
			Username: etcdUserName,
			Password: etcdPasswd,
		}

		var err error
		etcdClient, err = NewClient(etcdServer, 50*time.Second, nil, acfg)
		if err != nil {
			log.Warn("connect etcd server failed!", err)
			return nil
		}
	}
	return etcdClient
}

func GetEtcdServiceList(prefixKey string) map[string]interface{} {
	etcdcli := EtcdClient()
	if etcdcli == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(etcdctx, etcdTimeout*time.Second)
	defer cancel()
	grep, err := etcdcli.Get(ctx, prefixKey, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcdClient.Get service fail, err:%v, keyname:%s", err, prefixKey)
		return nil
	}

	result := make(map[string]interface{}, len(grep.Kvs))
	for _, kv := range grep.Kvs {
		ips := strings.Split(string(kv.Key), "/")
		url := ips[len(ips)-1]

		log.Debugf("%s", string(kv.Value))

		var m map[string]interface{}
		json.Unmarshal(kv.Value, &m)
		if v, ok := m["service_host"].(string); ok {
			result[url] = v
		} else {
			result[url] = ""
		}
	}
	return result
}

func EtcdWatch(keyName string, isRoop bool, cb func(t int, k string, v map[string]interface{})) {
	log.Debug("add watching:", keyName)
	go func() {
		etcdcli := EtcdClient()
		for {
			wc := etcdcli.Watch(context.Background(), keyName, clientv3.WithPrefix())
			da := <-wc
			for _, change := range da.Events {
				log.Infof("etcd tell change,%+v", change)

				kvk := string(change.Kv.Key)
				var val map[string]interface{}
				json.Unmarshal(change.Kv.Value, &val)

				cb(int(change.Type), kvk, val)
				if !isRoop {
					return
				}
			}
		}
	}()
}

func newRpcxClients(sname string) map[string]*rpcx.Client {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	etcdConfs := GetEtcdServiceList(sname)
	if etcdConfs == nil || len(etcdConfs) == 0 {
		return nil
	}

	clis := make(map[string]*rpcx.Client, 0)
	for url, _ := range etcdConfs {
		if len(url) <= 10 {
			continue
		}
		client := rpcx.NewClient(&rpcx.DirectClientSelector{
			Network:     "tcp",
			Address:     url,
			DialTimeout: 20 * time.Second,
		})
		if client != nil {
			clis[path.Join(sname, url)] = client
		}
	}

	return clis
}

func newRpcxClient(sname string, cb func(t int, k string, v map[string]interface{})) *rpcx.Client {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ips := strings.Split(sname, "/")
	url := ips[len(ips)-1]
	client := rpcx.NewClient(&rpcx.DirectClientSelector{
		Network:     "tcp",
		Address:     url,
		DialTimeout: 20 * time.Second,
	})
	return client
}
