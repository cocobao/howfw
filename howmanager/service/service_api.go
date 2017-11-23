package service

import "github.com/cocobao/howfw/mode"

type Trans int

func (t *Trans) TransIn(req *mode.TransData, reply *mode.TransResp) (err error) {
	defer func() {
		err = nil
		reply.Code = 200
		reply.Err = ""
	}()

	return
}
