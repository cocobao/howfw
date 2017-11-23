package rpcser

import (
	"crypto/tls"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

type SecureCfg struct {
	Cert   string
	Key    string
	Cacert string

	InsecureTransport  bool
	InsecureSkipVerify bool
}

type AuthCfg struct {
	Username string
	Password string
}

//创建etcd客户端
func NewClient(endpoints []string, dialTimeout time.Duration, scfg *SecureCfg, acfg *AuthCfg) (*clientv3.Client, error) {
	cfg, err := newClientCfg(endpoints, dialTimeout, scfg, acfg)
	if err != nil {
		return nil, err
	}
	client, err := clientv3.New(*cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newClientCfg(endpoints []string, dialTimeout time.Duration, scfg *SecureCfg, acfg *AuthCfg) (*clientv3.Config, error) {
	cfg := &clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout * time.Second,
	}
	if acfg != nil {
		cfg.Username = acfg.Username
		cfg.Password = acfg.Password
	}

	if scfg != nil {
		var cfgtls *transport.TLSInfo
		tlsinfo := transport.TLSInfo{}
		if scfg.Cert != "" {
			tlsinfo.CertFile = scfg.Cert
			cfgtls = &tlsinfo
		}

		if scfg.Key != "" {
			tlsinfo.KeyFile = scfg.Key
			cfgtls = &tlsinfo
		}
		if scfg.Cacert != "" {
			tlsinfo.CAFile = scfg.Cacert
			cfgtls = &tlsinfo
		}
		if cfgtls != nil {
			clientTLS, err := cfgtls.ClientConfig()
			if err != nil {
				return nil, err
			}
			cfg.TLS = clientTLS
		}
		// if key/cert is not given but user wants secure connection, we
		// should still setup an empty tls configuration for gRPC to setup
		// secure connection.
		if cfg.TLS == nil && !scfg.InsecureTransport {
			cfg.TLS = &tls.Config{}
		}

		// If the user wants to skip TLS verification then we should set
		// the InsecureSkipVerify flag in tls configuration.
		if scfg.InsecureSkipVerify && cfg.TLS != nil {
			cfg.TLS.InsecureSkipVerify = true
		}

	}
	return cfg, nil
}
