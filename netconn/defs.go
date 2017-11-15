package netconn

import (
	"context"
	"fmt"
	"time"
)

const (
	messageCtx string = "message"
	serverCtx  string = "netconn"
	netIDCtx   string = "netid"
)

const MessageMaxBytes = 1 << 23 // 8M

type ErrUndefined int32

func (e ErrUndefined) Error() string {
	return fmt.Sprintf("undefined message type %d", e)
}

const (
	MaxConnections    = 1000
	BufferSize128     = 128
	BufferSize256     = 256
	BufferSize512     = 512
	BufferSize1024    = 1024
	defaultWorkersNum = 20
)

type onConnectFunc func(WriteCloser) bool
type onMessageFunc func([]byte, WriteCloser)
type onCloseFunc func(WriteCloser)
type onErrorFunc func(WriteCloser)

type OnTimeOut struct {
	Callback func(time.Time, WriteCloser)
	Ctx      context.Context
}

type WriteCloser interface {
	Write([]byte) error
	Close()
}
