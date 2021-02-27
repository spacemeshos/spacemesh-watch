package alert

import (
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spacemeshos/spacemesh-watch/config"
)

func Raise(message string) {
	api := slack.New(config.SlackAPIToken)
	_, _, err := api.PostMessage(config.SlackChannelName, slack.MsgOptionText("*Spacemesh Watch*: "+message, false))

	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"slackMessage": message,
		}).Error("error sending message to slack")
	}
}
