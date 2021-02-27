package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spacemeshos/spacemesh-watch/config"
	"github.com/spacemeshos/spacemesh-watch/monitor"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spacemesh-node-monitor",
	Short: "A tool to monitor and raise alerts for spacemesh nodes",
	Run: func(cmd *cobra.Command, args []string) {
		monitor.StartMonitoring()
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringSliceVar(&config.Nodes, "nodes", []string{}, "comma seperated list of node GRPC URLs")
	rootCmd.PersistentFlags().IntVarP(&config.LayerWaitTime, "layer-wait-time", "", 3600, "time in seconds to wait for verified layer to increment")

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
