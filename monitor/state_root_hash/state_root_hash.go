package state_root_hash

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/spacemesh-watch/alert"
	"github.com/spacemeshos/spacemesh-watch/config"
	"google.golang.org/grpc"
)

var hashes = map[string][]string{}
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

	hash := hex.EncodeToString(r.Response.RootHash)
	layer := fmt.Sprint(r.Response.Layer.Number)

	if !contains(hashes[layer], hash) {
		hashes[layer] = append(hashes[layer], hash)
	}
}

func scanNetwork() {
	for _, node := range config.Nodes {
		wg.Add(1)
		go scanNode(node)
	}

	wg.Wait()

	for key, layer := range hashes {
		if len(layer) > 1 {
			go alert.Raise("total "+strconv.Itoa(len(layer))+" forks in the network for layer "+key, "", "FORKS_SUMMARY")
			log.Error("total " + strconv.Itoa(len(layer)) + " forks in the network for layer " + key)
		}
	}

	hashes = map[string][]string{}
}

func MonitorStateRootHash() {
	log.Debug("Started monitoring global state root hash")
	scanNetwork()
	for range time.Tick(time.Duration(config.RootHashWaitTime) * time.Second) {
		scanNetwork()
	}
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
