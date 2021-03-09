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

func scanNode(address string) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching recent verified layer")

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
	layer, ok := layers[address]

	if ok == false {
		layers[address] = r.Status.VerifiedLayer.Number
		log.WithFields(log.Fields{
			"node":  address,
			"layer": r.Status.VerifiedLayer.Number,
		}).Debug("set initial verified layer")
	} else {
		if r.Status.VerifiedLayer.Number <= layer {
			go alert.Raise("verified layer is stuck. current verified layer: "+strconv.FormatUint(uint64(layer), 10), address, "")
			log.WithFields(log.Fields{
				"node":  address,
				"layer": layer,
			}).Error("verified layer is stuck")
		} else {
			layers[address] = r.Status.VerifiedLayer.Number
			log.WithFields(log.Fields{
				"node":  address,
				"layer": layer,
			}).Debug("verified layer incremented")
		}
	}
}

func scanNetwork() {
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()
}

func MonitorVerifiedLayerProgress() {
	log.Debug("Started monitoring verified layer progress")
	scanNetwork()
	for range time.Tick(time.Duration(config.LayerWaitTime) * time.Second) {
		scanNetwork()
	}
}
