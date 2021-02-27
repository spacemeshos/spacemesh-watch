package verified_layer

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/spacemesh-watch/config"
	"google.golang.org/grpc"
)

var layers = make(map[string]uint32)
var wg sync.WaitGroup

func scanNode(address string) {
	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching recent verified layer")

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.WithFields(log.Fields{
			"node":  address,
			"error": err,
		}).Error("could not connect to API service")
	}

	defer conn.Close()

	c := pb.NewNodeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Status(ctx, &pb.StatusRequest{})

	if err != nil {
		log.WithFields(log.Fields{
			"node":  address,
			"error": err,
		}).Error("could not fetch status")
	}

	layer, ok := layers[address]

	if ok == false {
		layers[address] = r.Status.VerifiedLayer.Number
		log.WithFields(log.Fields{
			"node":  address,
			"layer": r.Status.VerifiedLayer.Number,
		}).Debug("set initial verified layer")
	} else {
		if r.Status.VerifiedLayer.Number <= layer {
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

	defer wg.Done()
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
