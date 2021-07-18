package config

import (
	"strings"
)

var (
	Nodes                      []string
	LayerWaitTime              int
	SyncWaitTime               int
	RootHashWaitTime           int
	SlackAPIToken              string
	SlackChannelName           string
	SlackMessageLimit          int
	SlackMessageLimitResetTime int
	NodeNames                  map[string]string
)

func Init() {
	urls := []string{}

	NodeNames = make(map[string]string)

	for _, node := range Nodes {
		s := strings.Split(node, "/")
		urls = append(urls, s[0])

		if len(s) == 2 {
			NodeNames[s[0]] = s[1]
		} else {
			NodeNames[s[0]] = s[0]
		}
	}

	Nodes = urls
}
