package netconn

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cocobao/log"
	"spm.pub/cloud/deev/utils"
)

func readLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		rawConn          net.Conn
		cDone            <-chan struct{}
		sDone            <-chan struct{}
		setHeartBeatFunc func(int64)
		onMessage        onMessageFunc
	)

	switch c := c.(type) {
	case *ServerConn:
		rawConn = c.rawConn
		cDone = c.ctx.Done()
		sDone = c.belong.ctx.Done()
		setHeartBeatFunc = c.SetHeartBeat
		onMessage = c.belong.opts.onMessage

	case *ClientConn:
		rawConn = c.rawConn
		cDone = c.ctx.Done()
		sDone = nil
		setHeartBeatFunc = c.SetHeartBeat
		onMessage = c.opts.onMessage
	}

	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v\n", p)
			utils.PrintStack()
		}
		wg.Done()
		log.Debug("readLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case <-cDone: // connection closed
			log.Debug("receiving cancel signal from conn")
			return
		case <-sDone: // netconn closed
			log.Debug("receiving cancel signal from netconn")
			return
		default:
			lengthBytes := make([]byte, 4)
			_, err := io.ReadFull(rawConn, lengthBytes)
			if err != nil {
				if io.EOF == err {
					log.Warn("conn has been close by peer")
					return
				}
				log.Warn("read length bytes fail", err)
				return
			}

			if len(lengthBytes) == 0 {
				log.Warn("length bytes is 0")
				continue
			}

			lengthBuf := bytes.NewReader(lengthBytes)
			var msgLen uint32
			if err = binary.Read(lengthBuf, binary.LittleEndian, &msgLen); err != nil {
				log.Warn("lengthBuf read fail", err)
				return
			}

			if msgLen > MessageMaxBytes {
				log.Warn("msg leng invalid,", msgLen)
				return
			}

			msgBytes := make([]byte, msgLen)
			_, err = io.ReadFull(rawConn, msgBytes)
			if err != nil {
				log.Warn("io read fail", err)
				return
			}
			setHeartBeatFunc(time.Now().UnixNano())
			onMessage(msgBytes, c.(WriteCloser))
		}
	}
}

func writeLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		rawConn net.Conn
		sendCh  chan []byte
		cDone   <-chan struct{}
		sDone   <-chan struct{}
		pkt     []byte
		err     error
	)

	switch c := c.(type) {
	case *ServerConn:
		rawConn = c.rawConn
		sendCh = c.sendCh
		cDone = c.ctx.Done()
		sDone = c.belong.ctx.Done()
	case *ClientConn:
		rawConn = c.rawConn
		sendCh = c.sendCh
		cDone = c.ctx.Done()
		sDone = nil
	}

	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panics: %v\n", p)
			utils.PrintStack()
		}
	OuterFor:
		for {
			select {
			case pkt = <-sendCh:
				if pkt != nil {
					if _, err = rawConn.Write(pkt); err != nil {
						log.Errorf("error writing data %v\n", err)
					}
				}
			default:
				break OuterFor
			}
		}
		wg.Done()
		log.Debug("writeLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case <-cDone: // connection closed
			log.Debug("receiving cancel signal from conn")
			return
		case <-sDone: // netconn closed
			log.Debug("receiving cancel signal from netconn")
			return
		case pkt = <-sendCh:
			if pkt != nil {
				//数据发送
				if _, err = rawConn.Write(pkt); err != nil {
					log.Errorf("error writing data %v\n", err)
					return
				}
			}
		}
	}
}

// func handleLoop(c WriteCloser, wg *sync.WaitGroup) {
// 	var (
// 		cDone   <-chan struct{}
// 		sDone   <-chan struct{}
// 		timerCh chan *OnTimeOut
// handlerCh chan MessageHandler
// netID        int64
// ctx          context.Context
// askForWorker bool
// err          error
// )

// switch c := c.(type) {
// case *ServerConn:
// 	cDone = c.ctx.Done()
// 	sDone = c.belong.ctx.Done()
// 	timerCh = c.timerCh
// handlerCh = c.handlerCh
// netID = c.netid
// ctx = c.ctx
// askForWorker = true
// case *ClientConn:
// 	cDone = c.ctx.Done()
// 	sDone = nil
// 	timerCh = c.timing.timeOutChan
// 	handlerCh = c.handlerCh
// 	netID = c.netid
// 	ctx = c.ctx
// }

// defer func() {
// 	if p := recover(); p != nil {
// 		log.Errorf("panics: %v\n", p)
// 	}
// 	wg.Done()
// 	log.Debug("handleLoop go-routine exited")
// 	c.Close()
// }()

// for {
// 	select {
// 	case <-cDone: // connectin closed
// 		log.Debug("receiving cancel signal from conn")
// 		return
// 	case <-sDone: // netconn closed
// 		log.Debug("receiving cancel signal from netconn")
// 		return
// case msgHandler := <-handlerCh:
// 	msg, handler := msgHandler.message, msgHandler.handler
// 	if handler != nil {
// 		log.Debug(msg)
// if askForWorker {
// 	err = WorkerPoolInstance().Put(netID, func() {
// 		handler(NewContextWithNetID(NewContextWithMessage(ctx, msg), netID), c)
// 	})
// 	if err != nil {
// 		log.Error(err)
// 	}
// } else {
// 	handler(NewContextWithNetID(NewContextWithMessage(ctx, msg), netID), c)
// }
// }
// case timeout := <-timerCh:
// 	if timeout != nil {
// timeoutNetID := NetIDFromContext(timeout.Ctx)
// if timeoutNetID != netID {
// 	log.Errorf("timeout net %d, conn net %d, mismatched!\n", timeoutNetID, netID)
// }
// if askForWorker {
// 	err = WorkerPoolInstance().Put(netID, func() {
// 		timeout.Callback(time.Now(), c.(WriteCloser))
// 	})
// 	if err != nil {
// 		log.Error(err)
// 	}
// } else {
// 	timeout.Callback(time.Now(), c.(WriteCloser))
// }
// 			}
// 		}
// 	}
// }
