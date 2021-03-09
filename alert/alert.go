package alert

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spacemeshos/spacemesh-watch/config"
)

var alertsTracker = make(map[string]map[string]bool)
var mu sync.Mutex

func Raise(message string, miner string, msg_type string) {
	api := slack.New(config.SlackAPIToken)

	if msg_type != "" {
		mu.Lock()
		defer mu.Unlock()
		minerMap, ok := alertsTracker[miner]

		if ok == false {
			alertsTracker[miner] = make(map[string]bool)
			alertsTracker[miner][msg_type] = true
		} else if minerMap[msg_type] == false {
			minerMap[msg_type] = true
		} else if minerMap[msg_type] == true {
			return
		}
	}

	_, _, err := api.PostMessage(config.SlackChannelName, slack.MsgOptionText("*Spacemesh Watch (Miner: "+miner+")*: "+message, false))

	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"slackMessage": message,
		}).Error("error sending message to slack")
	}
}
