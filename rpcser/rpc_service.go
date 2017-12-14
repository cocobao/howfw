package rpcser

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	"sync"

	"github.com/cocobao/log"
	"github.com/coreos/etcd/clientv3"
	"github.com/smallnest/rpcx"
)

var (
	etcdClient *clientv3.Client
)

type ServiceClisInfo struct {
	Client  *rpcx.Client
	MapData map[string]interface{}
	DevList []string
}

type RpcxClis struct {
	PrefixName    string
	ClientsMap    map[string]*ServiceClisInfo
	OnConnectCall func(cliKey string)
	Lock          sync.Mutex
}

func (r *RpcxClis) AddDev(host string, did string) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	k := path.Join(r.PrefixName, host)
	if v, ok := r.ClientsMap[k]; ok {
		for _, v := range v.DevList {
			if v == did {
				return
			}
		}
		log.Debugf("add dev, service:%s, did:%s", host, did)
		v.DevList = append(v.DevList, did)
	}
}

func (r *RpcxClis) AddDevWithKey(key string, did string) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if v, ok := r.ClientsMap[key]; ok {
		for _, v := range v.DevList {
			if v == did {
				return
			}
		}
		log.Debugf("add dev, service:%s, did:%s", key, did)
		v.DevList = append(v.DevList, did)
	}
}

func (r *RpcxClis) DecDev(host string, did string) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	k := path.Join(r.PrefixName, host)
	if v, ok := r.ClientsMap[k]; ok {
		log.Debugf("dec dev, service:%s, did:%s", host, did)
		for i, devid := range v.DevList {
			if devid == did {
				v.DevList = append(v.DevList[:i], v.DevList[i+1:]...)
				return
			}
		}
	}
}

func (r *RpcxClis) DevBelongTo(did string) string {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	for k, v := range r.ClientsMap {
		for _, dev := range v.DevList {
			if dev == did {
				hs := strings.Split(k, "/")
				return hs[len(hs)-1]
			}
		}
	}
	return ""
}

func (r *RpcxClis) SendDataToClient(host string, method string, args interface{}, reply interface{}) {
	r.Lock.Lock()
	k := path.Join(r.PrefixName, host)
	if v, ok := r.ClientsMap[k]; ok {
		r.Lock.Unlock()
		ctx, cancel := context.WithTimeout(etcdctx, etcdTimeout*time.Second)
		v.Client.Call(ctx, method, args, reply)
		cancel()
	} else {
		log.Warnf("not found:%s", k)
	}
}

func (r *RpcxClis) ApplyServiceHost() string {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if r.ClientsMap == nil || len(r.ClientsMap) == 0 {
		return ""
	}

	less := math.MaxUint32
	kk := ""
	for k, v := range r.ClientsMap {
		l := len(v.DevList)
		if l < less {
			less = l
			kk = k
		}
	}

	if v, ok := r.ClientsMap[kk]; ok {
		if sh, ok := v.MapData["service_host"].(string); ok {
			return sh
		}
	}

	return ""
}

func (r *RpcxClis) LoadClients() {
	if r.ClientsMap == nil {
		r.ClientsMap = make(map[string]*ServiceClisInfo, 0)

		mc, md := newRpcxClients(r.PrefixName)
		for k, v := range mc {
			sci := &ServiceClisInfo{
				Client:  v,
				MapData: md[k],
				DevList: make([]string, 0),
			}
			r.ClientsMap[k] = sci

			if r.OnConnectCall != nil {
				r.OnConnectCall(k)
			}
		}

		EtcdWatch(r.PrefixName, true, r.changeCalls)
	}
}

func (r *RpcxClis) CallClient(key string, method string, args interface{}, reply interface{}) {
	r.Lock.Lock()
	if cli, ok := r.ClientsMap[key]; ok {
		r.Lock.Unlock()
		ctx, cancel := context.WithTimeout(etcdctx, etcdTimeout*time.Second)
		defer cancel()
		cli.Client.Call(ctx, method, args, reply)
	}
}

func (r *RpcxClis) changeCalls(t int, k string, v map[string]interface{}) {
	if t == 0 {
		c := newRpcxClient(k, r.changeCalls)
		sci := &ServiceClisInfo{
			Client:  c,
			MapData: v,
			DevList: []string{},
		}
		r.Lock.Lock()
		r.ClientsMap[k] = sci
		r.Lock.Unlock()

		if r.OnConnectCall != nil {
			r.OnConnectCall(k)
		}
		log.Debug("new rpcx client,", k, v)
	} else if t == 1 {
		r.Lock.Lock()
		defer r.Lock.Unlock()
		if vc, ok := r.ClientsMap[k]; ok {
			vc.Client.Close()
			delete(r.ClientsMap, k)
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

func GetEtcdServiceList(prefixKey string) map[string]map[string]interface{} {
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

	result := make(map[string]map[string]interface{}, len(grep.Kvs))
	for _, kv := range grep.Kvs {
		ips := strings.Split(string(kv.Key), "/")
		url := ips[len(ips)-1]

		log.Debugf("%s", string(kv.Value))

		var m map[string]interface{}
		json.Unmarshal(kv.Value, &m)
		result[url] = m
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

func newRpcxClients(sname string) (map[string]*rpcx.Client, map[string]map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	etcdConfs := GetEtcdServiceList(sname)
	if etcdConfs == nil || len(etcdConfs) == 0 {
		return nil, nil
	}

	clis := make(map[string]*rpcx.Client, 0)
	md := make(map[string]map[string]interface{}, 0)
	for url, m := range etcdConfs {
		if len(url) <= 10 {
			continue
		}
		client := rpcx.NewClient(&rpcx.DirectClientSelector{
			Network:     "tcp",
			Address:     url,
			DialTimeout: 20 * time.Second,
		})
		if client != nil {
			k := path.Join(sname, url)
			clis[k] = client
			md[k] = m
			log.Debug("new client:", k)
		}
	}

	return clis, md
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
