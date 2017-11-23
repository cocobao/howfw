package netconn

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cocobao/howfw/util"
	"github.com/cocobao/log"
)

type ClientConn struct {
	addr    string
	opts    options
	netid   int64
	rawConn net.Conn
	once    *sync.Once
	wg      *sync.WaitGroup
	sendCh  chan []byte

	mu      sync.Mutex
	name    string
	heart   int64
	pending []int64
	ctx     context.Context
	cancel  context.CancelFunc
}

//新建客户端
func NewClientConn(netid int64, addr string, opt ...ServerOption) *ClientConn {
	var opts options
	for _, o := range opt {
		o(&opts)
	}

	if opts.bufferSize <= 0 {
		opts.bufferSize = BufferSize256
	}
	return newClientConnWithOptions(netid, addr, opts)
}

func newClientConnWithOptions(netid int64, addr string, opts options) *ClientConn {
	var c net.Conn
	var err error
	if opts.tlsCfg != nil {
		c, err = tls.Dial("tcp", addr, opts.tlsCfg)
	} else {
		c, err = net.Dial("tcp", addr)
	}

	if err != nil {
		log.Error(err)
		os.Exit(0)
	}

	cc := &ClientConn{
		addr:    c.RemoteAddr().String(),
		opts:    opts,
		netid:   netid,
		rawConn: c,
		once:    &sync.Once{},
		wg:      &sync.WaitGroup{},
		sendCh:  make(chan []byte, opts.bufferSize),

		heart: time.Now().UnixNano(),
	}
	cc.ctx, cc.cancel = context.WithCancel(context.Background())
	cc.name = c.RemoteAddr().String()
	cc.pending = []int64{}
	return cc
}

func (cc *ClientConn) SetHeartBeat(heart int64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.heart = heart
}

func (cc *ClientConn) HeartBeat() int64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	heart := cc.heart
	return heart
}

func (cc *ClientConn) Write(message []byte) error {
	return asyncWrite(cc, message)
}

func (cc *ClientConn) Start() {
	log.Infof("conn start, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())
	onConnect := cc.opts.onConnect
	if onConnect != nil {
		onConnect(cc)
	}

	loopers := []func(WriteCloser, *sync.WaitGroup){readLoop, writeLoop}
	for _, l := range loopers {
		looper := l
		cc.wg.Add(1)
		go looper(cc, cc.wg)
	}
}

func (cc *ClientConn) Close() {
	cc.once.Do(func() {
		log.Infof("conn close gracefully, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())

		onClose := cc.opts.onClose
		if onClose != nil {
			onClose(cc)
		}

		cc.rawConn.Close()

		cc.mu.Lock()
		cc.cancel()
		cc.pending = nil
		cc.mu.Unlock()

		cc.wg.Wait()

		close(cc.sendCh)

		if cc.opts.reconnect {
			util.Wait(5)
			cc.reconnect()
		}
	})
}

func (cc *ClientConn) reconnect() {
	*cc = *newClientConnWithOptions(cc.netid, cc.addr, cc.opts)
	cc.Start()
}
