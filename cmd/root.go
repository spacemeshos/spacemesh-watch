package cmd

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spacemeshos/spacemesh-watch/config"
	"github.com/spacemeshos/spacemesh-watch/monitor"
	"github.com/spacemeshos/spacemesh-watch/alert"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spacemesh-node-monitor",
	Short: "A tool to monitor and raise alerts for spacemesh nodes",
	Run: func(cmd *cobra.Command, args []string) {
		if len(config.Nodes) > 0 {
			go alert.ResetTracker()
			go monitor.StartMonitoring()

			ctx, _ := context.WithCancel(context.Background())
			<-ctx.Done()
		} else {
			log.Info("No nodes configured for monitoring. Exiting service...")
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringSliceVar(&config.Nodes, "nodes", []string{}, "comma seperated list of node GRPC URLs")
	rootCmd.PersistentFlags().IntVarP(&config.LayerWaitTime, "layer-wait-time", "", 660, "time in seconds to wait for verified layer to increment")
	rootCmd.PersistentFlags().IntVarP(&config.SyncWaitTime, "sync-wait-time", "", 3600, "time in seconds to wait node status to change from syncing to synced after first")
	rootCmd.PersistentFlags().IntVarP(&config.RootHashWaitTime, "state-root-hash-wait-time", "", 600, "time in seconds to wait for checking if state root hash is same for all nodes for the same layer number")
	rootCmd.PersistentFlags().StringVarP(&config.SlackAPIToken, "slack-api-token", "", "", "slack API token for authorizing assert notifications. create a slack app and generate \"Bot User OAuth Token\" and set set chat:write, chat:write.public and im:write permissions")
	rootCmd.PersistentFlags().StringVarP(&config.SlackChannelName, "slack-channel-name", "", "", "slack channel name to send messages to. its the last path in the browser URL for the channel")
	rootCmd.PersistentFlags().IntVarP(&config.SlackMessageLimit, "slack-message-limit", "", 4, "number of messages of a alert type after which its silenced")
	rootCmd.PersistentFlags().IntVarP(&config.SlackMessageLimitResetTime, "slack-message-limit-reset", "", 3600, "number of seconds after which slack message limit is reset")

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
