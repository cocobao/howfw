package rpcser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cocobao/howfw/util/timeutil"
	"github.com/cocobao/log"
	"github.com/smallnest/rpcx"
	rpcxlog "github.com/smallnest/rpcx/log"
	emptylog "github.com/ti/goutil/log"
)

var (
	//rpcx服务对象
	rpcxServer *rpcx.Server
	//本地ip和端口
	localHost string
	//rpcx服务名称
	serviceName string

	//etcd相关信息
	etcdServer   []string
	etcdUserName string
	etcdPasswd   string
	etcdTimeout  time.Duration
	etcdctx      context.Context
)

func SetupEtcdConfig(host, usname, pwd, sn string,
	endPoints []string,
	dialTimeout time.Duration) {

	localHost = host
	etcdServer = endPoints
	etcdUserName = usname
	etcdPasswd = pwd
	etcdTimeout = dialTimeout
	serviceName = sn
	etcdctx = context.Background()
}

//配置etcd相关配置
func SetupRpcx(service interface{}, serviceHost string) error {
	etcdCli := EtcdClient()
	if etcdCli == nil {
		return fmt.Errorf("connect etcd server failed!")
	} else {

		jm, _ := json.Marshal(map[string]string{
			"create_at":    timeutil.TimeToZoneStr(time.Now().Unix()),
			"service_host": serviceHost,
		})

		key := "/" + serviceName + "/" + localHost
		ctx, cancel := context.WithTimeout(etcdctx, etcdTimeout*time.Second)
		_, err := etcdCli.Put(ctx, key, string(jm))
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		cancel()
	}

	etcdURLs := "etcd://" + etcdUserName + ":" + etcdPasswd + "@"
	for i, v := range etcdServer {
		etcdURLs += strings.Split(v, "//")[1]
		if i < len(etcdServer)-1 {
			etcdURLs += ","
		}
	}

	rpcxServer = rpcx.NewServer()
	rpcxServer.RegisterName(serviceName, service, "")
	rpcxlog.SetLogger(emptylog.NewDefaultLogger(emptylog.EmpWriter()))

	log.Info("setup etcd success")
	return nil
}

//运行rpcx服务
func RunRpcServer() {
	err := rpcxServer.Serve("tcp", localHost)
	if err != nil {
		fmt.Println("rpcxServer.Serve run fail", err)
		os.Exit(0)
	}
}

//注销服务,以及注销在etcd的配置
func StopRpcServer() {
	rpcxServer.Close()

	acfg := &AuthCfg{
		Username: etcdUserName,
		Password: etcdPasswd,
	}

	etcdClient, err := NewClient(etcdServer, 50*time.Second, nil, acfg)
	if err != nil {
		log.Warn("connect etcd server failed!", err)
		return
	}
	ctx, cancel := context.WithTimeout(etcdctx, etcdTimeout*time.Second)
	etcdClient.Delete(ctx, "/"+serviceName+"/"+localHost)
	etcdClient.Close()
	cancel()
}
