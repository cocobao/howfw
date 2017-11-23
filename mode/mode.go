package mode

import "github.com/cocobao/howfw/netconn"

type Clipoint struct {
	Nid   int64
	Name  string
	Addr  string
	Conn  netconn.WriteCloser
	Binds []string
}

type TransData struct {
	Headers map[string]string `msg:header`
	Body    []byte            `msg:"body"`
}

type TransResp struct {
	RespData TransData `msg:"args"`
	Err      string    `msg:"errmsg"`
	Code     int64     `msg:"code"`
}
