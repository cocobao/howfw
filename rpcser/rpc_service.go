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
	//etcd客户端
	etcdClient *clientv3.Client
)

//服务客户端信息
type ServiceClisInfo struct {
	//rpcx客户端连续
	Client  *rpcx.Client
	MapData map[string]interface{}

	//所带设备端列表
	DevList []string
}

//rpcx客户端数据结构
type RpcxClis struct {
	//名称前缀
	PrefixName string

	//rpcx客户端列表
	ClientsMap map[string]*ServiceClisInfo

	//有新连接回调
	OnConnectCall func(cliKey string)

	Lock sync.Mutex
}

//添加一个设备到到指定ip的service客户端
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

//添加一个设备到一个指定的serice客户端里
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

//从一个指定ip的service客户端里删除一个设备
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

//查看一个设备归属的service客户端
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

//调用rpcx客户端服务接口发送数据
func (r *RpcxClis) SendDataToClient(host string, method string, args interface{}, reply interface{}) {
	k := path.Join(r.PrefixName, host)
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if v, ok := r.ClientsMap[k]; ok {
		ctx, cancel := context.WithTimeout(etcdctx, 5*time.Second)
		v.Client.Call(ctx, method, args, reply)
		cancel()
	} else {
		log.Warnf("not found:%s", k)
	}
}

//查找一个最少设备的服务节点
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

//加载rpcx客户端管理器
func (r *RpcxClis) LoadClients() {
	if r.ClientsMap == nil {
		r.ClientsMap = make(map[string]*ServiceClisInfo, 0)

		//预加载所有的rpcx客户端
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

		//侦听这个服务, 有可能会有节点新增或者删除
		EtcdWatch(r.PrefixName, true, r.changeCalls)
	}
}

func (r *RpcxClis) CallClient(key string, method string, args interface{}, reply interface{}) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if cli, ok := r.ClientsMap[key]; ok {
		ctx, cancel := context.WithTimeout(etcdctx, 5*time.Second)
		cli.Client.Call(ctx, method, args, reply)
		cancel()
	}
}

func (r *RpcxClis) changeCalls(t int, k string, v map[string]interface{}) {
	//有新的节点注册到etcd
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

		//有节点从etcd注销了
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

	//获取所有注册的服务端列表
	grep, err := etcdcli.Get(ctx, prefixKey, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcdClient.Get service fail, err:%v, keyname:%s", err, prefixKey)
		return nil
	}

	result := make(map[string]map[string]interface{}, len(grep.Kvs))
	for _, kv := range grep.Kvs {
		log.Debugf("%s", string(kv.Value))

		ips := strings.Split(string(kv.Key), "/")
		if len(ips) > 0 {
			//路径最后一段约定为ip:port形式
			url := ips[len(ips)-1]

			var m map[string]interface{}
			if err := json.Unmarshal(kv.Value, &m); err == nil {
				result[url] = m
			} else {
				log.Warn("unmarshal fail,", err)
			}
		}
	}
	return result
}

//侦听etcd
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

//新建一个rpcx客户端
func newRpcxClients(sname string) (map[string]*rpcx.Client, map[string]map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	//获取所有service客户端列表
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

		//简历rpcx连接
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

//新建一个rpcx客户端连接
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
