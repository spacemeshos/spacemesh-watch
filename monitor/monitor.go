package monitor

import (
	"github.com/spacemeshos/spacemesh-watch/monitor/verified_layer"
)

func StartMonitoring() {
	go verified_layer.MonitorVerifiedLayerProgress()
}
