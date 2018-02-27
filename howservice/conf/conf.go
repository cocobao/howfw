package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cocobao/log"
	yaml "gopkg.in/yaml.v2"
)

const (
	VERSION string = "3.02.01.11160946"
)

var (
	ver                 = flag.Bool("ver", false, "-- the version")
	SetConfPath *string = flag.String("conf", "conf/settings.yml", "-- config file")
	GCfg        *Config
)

type EtcdCfg struct {
	Endpoints   []string      `yaml:"endpoints"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Username    string        `yaml:"user_name"`
	Password    string        `yaml:"password"`
}

type Config struct {
	LocalHost   string  `yaml:"local_host"`
	ServiceHost string  `yaml:"service_host"`
	EtcdServer  EtcdCfg `yaml:"etcd_server"`
	MongoHost   string  `yaml:"mongo_host"`
	CerPath     string  `yaml:"cer_path"`
}

func NewConfig() (*Config, error) {
	flag.Parse()
	if *ver {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	cfg := &Config{}
	if err := cfg.configFromFile(*SetConfPath); err != nil {
		return nil, err
	}

	fmt.Printf("cfg:%+v\n", cfg)
	return cfg, nil
}

func (cfg *Config) configFromFile(path string) error {
	b, rerr := ioutil.ReadFile(path)
	if rerr != nil {
		return rerr
	}
	if yerr := yaml.Unmarshal(b, cfg); yerr != nil {
		return yerr
	}
	return nil
}

func init() {
	//读取配置文件
	var err error
	GCfg, err = NewConfig()
	if err != nil {
		fmt.Println(err)
		panic("read config file fail")
	}
	fmt.Println("read config file ok")

	log.NewLogger("", log.LoggerLevelDebug)
}
