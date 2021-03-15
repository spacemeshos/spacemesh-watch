package monitor

import (
	"github.com/spacemeshos/spacemesh-watch/monitor/sync_status"
	"github.com/spacemeshos/spacemesh-watch/monitor/verified_layer"
)

func StartMonitoring() {
	go verified_layer.MonitorVerifiedLayerProgress()
	go sync_status.MonitorSyncStatus()
}
