package monitor

import (
	"github.com/spacemeshos/spacemesh-watch/monitor/sync_status"
)

func StartMonitoring() {
	go verified_layer.MonitorVerifiedLayerProgress()
	go sync_status.MonitorSyncStatus()
}
