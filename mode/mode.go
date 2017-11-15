package mode

import "github.com/cocobao/howfw/netconn"

type Clipoint struct {
	Name  string
	Addr  string
	Conn  netconn.WriteCloser
	Binds []string
}
