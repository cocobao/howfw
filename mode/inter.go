package mode

type CallClimgr interface {
	SendDataToDev(devId string, data map[string]interface{})
	GetCliList() []string
}

type CallService interface {
	CallManager(md map[string]interface{})
}
