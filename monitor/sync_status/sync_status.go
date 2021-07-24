package sync_status

import (
	"context"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/spacemesh-watch/alert"
	"github.com/spacemeshos/spacemesh-watch/config"
	"google.golang.org/grpc"
)

type SyncData struct {
	isSynced    bool
	syncedLayer uint32
}

var syncStatus = make(map[string]*SyncData)
var wg sync.WaitGroup
var mu sync.Mutex
var totalStuckNodes = 0
var totalCrashedNodes = 0

func scanNode(address string) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching node status")

	conn, _ := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()

	c := pb.NewNodeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := c.Status(ctx, &pb.StatusRequest{})

	mu.Lock()
	defer mu.Unlock()

	if err != nil {
		go alert.Raise("could not fetch status. Error: "+err.Error(), address, "CONNECTION_ERROR")
		log.WithFields(log.Fields{
			"node":  address,
			"error": err.Error(),
		}).Error("could not fetch status")
		totalCrashedNodes++
		return
	}

	status, ok := syncStatus[address]

	if ok == false {
		syncStatus[address] = &SyncData{r.Status.IsSynced, r.Status.SyncedLayer.Number}
		log.WithFields(log.Fields{
			"node":   address,
			"synced": r.Status.IsSynced,
			"layer":  r.Status.SyncedLayer.Number,
		}).Info("set initial sync status")
	} else {
		if r.Status.IsSynced == false {
			if status.syncedLayer < r.Status.SyncedLayer.Number {
				log.WithFields(log.Fields{
					"node":  address,
					"layer": r.Status.SyncedLayer.Number,
				}).Info("node still syncing")
			} else {
				totalStuckNodes++
				go alert.Raise("node not syncing. current synced layer: "+strconv.FormatUint(uint64(status.syncedLayer), 10), address, "SYNC_STATUS")
				log.WithFields(log.Fields{
					"node":  address,
					"layer": r.Status.SyncedLayer.Number,
				}).Error("node stuck at syncing")
			}
		} else {
			log.WithFields(log.Fields{
				"node":  address,
				"layer": r.Status.SyncedLayer.Number,
			}).Info("node synced")
		}

		status.isSynced = r.Status.IsSynced
		status.syncedLayer = r.Status.SyncedLayer.Number
	}

}

func scanNetwork() {
	totalStuckNodes = 0
	totalCrashedNodes = 0
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()

	if totalStuckNodes != 0 {
		go alert.Raise("total "+strconv.Itoa(totalStuckNodes)+" nodes are stuck syncing", "", "VERIFIED_LAYER_SUMMARY")
		log.Error("total " + strconv.Itoa(totalStuckNodes) + " nodes are stuck syncing")
	}

	if totalCrashedNodes != 0 {
		go alert.Raise("total "+strconv.Itoa(totalCrashedNodes)+" nodes have crashed", "", "CRASHED_NODES_SUMMARY")
		log.Error("total " + strconv.Itoa(totalCrashedNodes) + " nodes have crashed")
	}
}

func MonitorSyncStatus() {
	log.Debug("Started monitoring sync status")
	scanNetwork()
	for range time.Tick(time.Duration(config.SyncWaitTime) * time.Second) {
		scanNetwork()
	}
}
