package state_root_hash

import (
	"encoding/hex"
	"context"
	"sync"
	"time"
	"strconv"

	log "github.com/sirupsen/logrus"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/spacemesh-watch/alert"
	"github.com/spacemeshos/spacemesh-watch/config"
	"google.golang.org/grpc"
)

type stateRootInfo struct {
  hash string
  layer uint32
	node string
}

var hashes = []stateRootInfo{}
var wg sync.WaitGroup
var mu sync.Mutex

func scanNode(address string) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"node": address,
	}).Debug("fetching state root hash")

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

	c := pb.NewGlobalStateServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := c.GlobalStateHash(ctx, &pb.GlobalStateHashRequest{})

	if err != nil {
		go alert.Raise("could not fetch state root hash. Error: "+err.Error(), address, "CONNECTION_ERROR")
		log.WithFields(log.Fields{
			"node":  address,
			"error": err.Error(),
		}).Error("could not fetch status")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	hashes = append(hashes, stateRootInfo{ hex.EncodeToString(r.Response.RootHash), r.Response.Layer.Number, address })
}

func compareHashes() {
	for ;; {
		if len(hashes) == 0 {
			break
		}

		info := &hashes[0]
		hashes = hashes[1:]

		for _, hash := range hashes {
			if hash.layer == info.layer {
				if hash.hash != info.hash {
					go alert.Raise("state root hash doesn't match for verified layer: "+strconv.FormatUint(uint64(info.layer), 10)+" when compared with node "+hash.node, info.node, "STATE_ROOT_HASH")
					log.WithFields(log.Fields{
						"node1":  hash.node,
						"node2": info.node,
						"hash1": hash.hash,
						"hash2": info.hash,
						"layer": hash.layer,
					}).Error("state root hash doesn't match")
				} else {
					log.WithFields(log.Fields{
						"node1":  hash.node,
						"node2": info.node,
						"hash": info.hash,
						"layer": hash.layer,
					}).Info("state root hash matches")
				}
			}
		}
	}

	hashes = []stateRootInfo{}
}

func scanNetwork() {
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()

	compareHashes()
}

func MonitorStateRootHash() {
	log.Debug("Started monitoring global state root hash")
	scanNetwork()
	for range time.Tick(time.Duration(config.RootHashWaitTime) * time.Second) {
		scanNetwork()
	}
}
