package climgr

import "github.com/cocobao/howfw/mode"

var (
	callService mode.CallService
)

func SetCallService(c mode.CallService) {
	callService = c
}

func SyncOnlineToManager(devId string) {
	callService.CallManager(true, map[string]interface{}{
		"cmd":   "dev_online",
		"devid": devId,
	})
}

func SyncOfflineToManager(devId string) {
	callService.CallManager(true, map[string]interface{}{
		"cmd":   "dev_offline",
		"devid": devId,
	})
}

func TransMsgToDev(from_id string, to_id string, data interface{}) {
	callService.CallManager(false, map[string]interface{}{
		"cmd":     "trans_data",
		"from_id": from_id,
		"to_id":   to_id,
		"data":    data,
	})
}
