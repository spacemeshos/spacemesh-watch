package alert

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spacemeshos/spacemesh-watch/config"
)

var alertsTracker = make(map[string]int)
var mu sync.Mutex

func Raise(message string, miner string, msg_type string) {
	api := slack.New(config.SlackAPIToken)

	mu.Lock()
	defer mu.Unlock()

	_, ok := alertsTracker[msg_type]

	sendMessage := true

	if ok == false {
		alertsTracker[msg_type] = 1
	} else if alertsTracker[msg_type] < config.SlackMessageLimit {
		alertsTracker[msg_type]++
	} else {
		sendMessage = false
	}

	if sendMessage == true {
		_, _, err := api.PostMessage(config.SlackChannelName, slack.MsgOptionText("*Spacemesh Watch (Miner: "+miner+")*: "+message, false))

		if err != nil {
			log.WithFields(log.Fields{
				"error":        err.Error(),
				"slackMessage": message,
			}).Error("error sending message to slack")
		}
	}
}

func ResetTracker() {
	for range time.Tick(time.Duration(config.SlackMessageLimitResetTime) * time.Second) {
		mu.Lock()
		defer mu.Unlock()
		alertsTracker = make(map[string]int)
	}
}