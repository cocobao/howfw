package mode

import "github.com/cocobao/howfw/netconn"

//客户端节点
type Clipoint struct {
	Nid  int64
	Name string
	Addr string
	Conn netconn.WriteCloser

	//客户端下的节点
	Binds []string

	//上线时间
	ConTime int64
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
