package netconn

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cocobao/log"
)

func init() {
	netIdentifier = NewAtomicInt64(1)
}

var (
	netIdentifier *AtomicInt64
	tlsWrapper    func(net.Conn) net.Conn
)

func IndId() int64 {
	return netIdentifier.GetAndIncrement()
}

func ReconnectOption() ServerOption {
	return func(o *options) {
		o.reconnect = true
	}
}

func OnConnectOption(cb func(WriteCloser) bool) ServerOption {
	return func(o *options) {
		o.onConnect = cb
	}
}

func OnMessageOption(cb func([]byte, WriteCloser)) ServerOption {
	return func(o *options) {
		o.onMessage = cb
	}
}

func OnCloseOption(cb func(WriteCloser)) ServerOption {
	return func(o *options) {
		o.onClose = cb
	}
}

func OnErrorOption(cb func(WriteCloser)) ServerOption {
	return func(o *options) {
		o.onError = cb
	}
}

func TLSCredsOption(config *tls.Config) ServerOption {
	return func(o *options) {
		o.tlsCfg = config
	}
}

type options struct {
	//tls配置
	tlsCfg *tls.Config
	//连接成功回调
	onConnect onConnectFunc
	//接收数据回调
	onMessage onMessageFunc
	//关闭连接回调
	onClose onCloseFunc
	//出错回调
	onError onErrorFunc

	workerSize int  // numbers of worker go-routines
	bufferSize int  // size of buffered channel
	reconnect  bool // for ClientConn use only
}

type ServerOption func(*options)

//新建服务
func NewServer(opt ...ServerOption) *Server {
	var opts options
	for _, o := range opt {
		o(&opts)
	}

	if opts.workerSize <= 0 {
		opts.workerSize = defaultWorkersNum
	}
	if opts.bufferSize <= 0 {
		opts.bufferSize = BufferSize256
	}

	s := &Server{
		opts:  opts,
		conns: &sync.Map{},
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

//获取连接
func (s *Server) Conn(id int64) (*ServerConn, bool) {
	v, ok := s.conns.Load(id)
	if ok {
		return v.(*ServerConn), ok
	}
	return nil, ok
}

type Server struct {
	opts  options
	conns *sync.Map
	//连接上下文
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Server) Start(port string) error {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		return fmt.Errorf("listen error", err)
	}
	log.Infof("server start, net %s addr %s\n", l.Addr().Network(), l.Addr().String())

	var tempDelay time.Duration
	for {
		rawConn, err := l.Accept()
		if err != nil {
			//accept出错，delay一下, 继续工作
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay >= max {
					tempDelay = max
				}
				log.Errorf("accept error %v, retrying in %d\n", err, tempDelay)
				select {
				case <-time.After(tempDelay):
				case <-s.ctx.Done():
				}
				continue
			}
			return err
		}
		tempDelay = 0

		if s.opts.tlsCfg != nil {
			rawConn = tls.Server(rawConn, s.opts.tlsCfg)
		}

		netid := IndId()
		sc := NewServerConn(netid, s, rawConn)
		s.conns.Store(netid, sc)

		go func() {
			sc.Start()
		}()

		log.Infof("accepted client %s, id %d\n", sc.Name(), netid)
		s.conns.Range(func(k, v interface{}) bool {
			i := k.(int64)
			c := v.(*ServerConn)
			log.Infof("client(%d) %s", i, c.Name())
			return true
		})
	}
}

func (s *Server) Stop() {
	conns := map[int64]*ServerConn{}

	s.conns.Range(func(k, v interface{}) bool {
		i := k.(int64)
		c := v.(*ServerConn)
		conns[i] = c
		return true
	})

	s.conns = nil

	for _, c := range conns {
		c.rawConn.Close()
		log.Infof("close client %s\n", c.Name())
	}

	log.Info("netconn stopped gracefully, bye.")
}
