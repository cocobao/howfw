package mode

type CallClimgr interface {
	SendDataToDev(devId string, data interface{})
	GetCliList() []string
}

type CallService interface {
	CallManager(md map[string]interface{})
}
