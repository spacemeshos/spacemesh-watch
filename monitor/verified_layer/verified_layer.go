package verified_layer

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

var layers = make(map[string]uint32)
var wg sync.WaitGroup
var mu sync.Mutex
var totalStuckNodes = 0

func scanNode(address string) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching recent verified layer")

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		go alert.Raise("could not connect to API server. Error: "+err.Error(), address, "CONNECTION_ERROR")
		log.WithFields(log.Fields{
			"node":  address,
			"error": err.Error(),
		}).Error("could not connect to API service")
		return
	}

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
	layer, ok := layers[address]

	if ok == false {
		layers[address] = r.Status.VerifiedLayer.Number
		log.WithFields(log.Fields{
			"node":  address,
			"layer": r.Status.VerifiedLayer.Number,
		}).Info("set initial verified layer")
		totalStuckNodes++
	} else {
		if r.Status.VerifiedLayer.Number <= layer {
			go alert.Raise("verified layer is stuck. current verified layer: "+strconv.FormatUint(uint64(layer), 10), address, "VERIFIED_LAYER")
			log.WithFields(log.Fields{
				"node":  address,
				"layer": layer,
			}).Error("verified layer is stuck")
		} else {
			layers[address] = r.Status.VerifiedLayer.Number
			log.WithFields(log.Fields{
				"node":  address,
				"layer": layer,
			}).Info("verified layer incremented")
		}
	}
}

func scanNetwork() {
	totalStuckNodes = 0
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()

	if totalStuckNodes != 0 {
		go alert.Raise("total "+strconv.Itoa(totalStuckNodes)+" nodes are stuck verifying layer", "", "VERIFIED_LAYER_SUMMARY")
		log.Error("total " + strconv.Itoa(totalStuckNodes) + " nodes are stuck verifying layer")
	}
}

func MonitorVerifiedLayerProgress() {
	log.Debug("Started monitoring verified layer progress")
	scanNetwork()
	for range time.Tick(time.Duration(config.LayerWaitTime) * time.Second) {
		scanNetwork()
	}
}
