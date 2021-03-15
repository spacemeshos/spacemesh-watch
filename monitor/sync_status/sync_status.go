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

func scanNode(address string) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching node status")

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		go alert.Raise("could not connect to API server. Error: "+err.Error(), address, "CONNECTION_ERROR")
		log.WithFields(log.Fields{
			"node":  address,
			"error": err.Error(),
		}).Error("could not connect to API service")
		return
	}

	defer conn.Close()

	c := pb.NewNodeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := c.Status(ctx, &pb.StatusRequest{})

	if err != nil {
		go alert.Raise("could not fetch status. Error: "+err.Error(), address, "CONNECTION_ERROR")
		log.WithFields(log.Fields{
			"node":  address,
			"error": err.Error(),
		}).Error("could not fetch status")
		return
	}

	mu.Lock()
	defer mu.Unlock()

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
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()
}

func MonitorSyncStatus() {
	log.Debug("Started monitoring sync status")
	scanNetwork()
	for range time.Tick(time.Duration(config.SyncWaitTime) * time.Second) {
		scanNetwork()
	}
}
