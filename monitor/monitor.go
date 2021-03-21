package monitor

import (
	"github.com/spacemeshos/spacemesh-watch/monitor/sync_status"
	"github.com/spacemeshos/spacemesh-watch/monitor/verified_layer"
	"github.com/spacemeshos/spacemesh-watch/monitor/state_root_hash"
)

func StartMonitoring() {
	go verified_layer.MonitorVerifiedLayerProgress()
	go sync_status.MonitorSyncStatus()
	go state_root_hash.MonitorStateRootHash()
}
