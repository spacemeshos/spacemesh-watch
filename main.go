package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spacemeshos/spacemesh-watch/cmd"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	cmd.Execute()

	ctx, _ := context.WithCancel(context.Background())
	<-ctx.Done()
}
