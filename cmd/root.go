package cmd

import (
	"context"
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

		ctx, _ := context.WithCancel(context.Background())
		<-ctx.Done()
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringSliceVar(&config.Nodes, "nodes", []string{}, "comma seperated list of node GRPC URLs")
	rootCmd.PersistentFlags().IntVarP(&config.LayerWaitTime, "layer-wait-time", "", 3600, "time in seconds to wait for verified layer to increment")
	rootCmd.PersistentFlags().StringVarP(&config.SlackAPIToken, "slack-api-token", "", "", "slack API token for authorizing assert notifications. create a slack app and generate user OAuth token and set set chat:write, chat:write.public and im:write permissions")
	rootCmd.PersistentFlags().StringVarP(&config.SlackChannelName, "slack-channel-name", "", "", "slack channel name to send messages to. its the last path in the browser URL for the channel")

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
