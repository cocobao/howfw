package mode

type CallClimgr interface {
	SendDataToDev(devId string, data map[string]interface{})
	GetCliList() []string
}

type CallService interface {
	CallManager(isBroadcast bool, md map[string]interface{})
}
