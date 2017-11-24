package climgr

import "github.com/cocobao/howfw/howservice/rpcservice"

func SyncOnlineToManager(devId string) {
	rpcservice.CallManager(map[string]interface{}{
		"cmd":   "dev_online",
		"devid": devId,
	})
}

func SyncOfflineToManager(devId string) {
	rpcservice.CallManager(map[string]interface{}{
		"cmd":   "dev_offline",
		"devid": devId,
	})
}
