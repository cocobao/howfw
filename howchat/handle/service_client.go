package handle

import (
	"github.com/cocobao/howfw/netconn"
	"github.com/cocobao/log"
)

func connectService(host string) {
	onConnect := netconn.OnConnectOption(func(c netconn.WriteCloser) bool {
		log.Info("on connect")
		return true
	})

	onError := netconn.OnErrorOption(func(c netconn.WriteCloser) {
		log.Info("on error")
	})

	onClose := netconn.OnCloseOption(func(c netconn.WriteCloser) {
		log.Info("on close")
		SetCon(nil)
	})

	onMessage := netconn.OnMessageOption(OnMessage)

	options := []netconn.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		// netconn.ReconnectOption(),
	}

	conn := netconn.NewClientConn(0, host, options...)
	conn.Start()
	SetCon(conn)

	// ctx := context.Background()
	// timerWheel := timer.NewTimingWheel(ctx)
	// timerWheel.AddTimer(netconn.IndId(), time.Now().Add(10*time.Second), 0, &timer.OnTimeOut{
	// 	Ctx:      ctx,
	// 	Callback: HeartBeat,
	// })
}
