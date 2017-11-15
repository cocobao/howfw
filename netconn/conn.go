package netconn

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cocobao/log"
	"spm.pub/cloud/deev/utils"
)

func NewServerConn(id int64, s *Server, c net.Conn) *ServerConn {
	sc := &ServerConn{
		netid:   id,
		belong:  s,
		rawConn: c,
		once:    &sync.Once{},
		wg:      &sync.WaitGroup{},
		sendCh:  make(chan []byte, s.opts.bufferSize),
		timerCh: make(chan *OnTimeOut, s.opts.bufferSize),
		heart:   time.Now().UnixNano(),
	}
	sc.ctx, sc.cancel = context.WithCancel(context.WithValue(s.ctx, serverCtx, s))
	sc.name = c.RemoteAddr().String()
	sc.pending = []int64{}
	return sc
}

type ServerConn struct {
	//内部识别id
	netid int64
	//对于的Server结构
	belong *Server
	//该连接
	rawConn net.Conn

	once *sync.Once
	wg   *sync.WaitGroup

	//发送队列
	sendCh chan []byte

	timerCh chan *OnTimeOut

	mu sync.Mutex
	//ip地址
	name string
	//心跳时间
	heart   int64
	pending []int64
	//连接上下文
	ctx context.Context
	//注销上下文函数
	cancel context.CancelFunc
}

func (sc *ServerConn) NetID() int64 {
	return sc.netid
}

func (sc *ServerConn) RemoteAddr() string {
	return sc.rawConn.RemoteAddr().String()
}

func (sc *ServerConn) Start() {
	log.Infof("conn start, <%v -> %v>\n", sc.rawConn.LocalAddr(), sc.rawConn.RemoteAddr())
	onConnect := sc.belong.opts.onConnect
	if onConnect != nil {
		onConnect(sc)
	}

	loopers := []func(WriteCloser, *sync.WaitGroup){readLoop, writeLoop}
	for _, l := range loopers {
		looper := l
		sc.wg.Add(1)
		go looper(sc, sc.wg)
	}
}

//关闭连接
func (sc *ServerConn) Close() {
	//连接关闭只能执行一次
	sc.once.Do(func() {
		log.Infof("conn close gracefully, <%v <- %v>\n", sc.rawConn.LocalAddr(), sc.rawConn.RemoteAddr())

		//回调onClose
		onClose := sc.belong.opts.onClose
		if onClose != nil {
			onClose(sc)
		}

		sc.belong.conns.Delete(sc.netid)

		//关闭网络连接
		if tc, ok := sc.rawConn.(*net.TCPConn); ok {
			tc.SetLinger(0)
		}
		sc.rawConn.Close()

		//cancel上下文
		sc.mu.Lock()
		sc.cancel()
		sc.mu.Unlock()

		sc.wg.Wait()

		close(sc.sendCh)
		close(sc.timerCh)
	})
}

func (sc *ServerConn) Write(message []byte) error {
	return asyncWrite(sc, message)
}

func (sc *ServerConn) Name() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	name := sc.name
	return name
}

//更新心跳时间
func (sc *ServerConn) SetHeartBeat(heart int64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.heart = heart
}

//读取最后心跳时间
func (sc *ServerConn) HeartBeat() int64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	heart := sc.heart
	return heart
}

func asyncWrite(c interface{}, m []byte) (err error) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Println(p)
			utils.PrintStack()
		}
	}()

	var (
		pkt    []byte
		sendCh chan []byte
	)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(len(m)))
	buf.Write(m)
	pkt = buf.Bytes()

	switch c := c.(type) {
	case *ServerConn:
		sendCh = c.sendCh

	case *ClientConn:
		sendCh = c.sendCh
	}

	if err != nil {
		log.Errorf("asyncWrite error %v\n", err)
		return
	}

	select {
	case sendCh <- pkt:
		err = nil
	default:
		err = fmt.Errorf("would block")
	}
	return
}
