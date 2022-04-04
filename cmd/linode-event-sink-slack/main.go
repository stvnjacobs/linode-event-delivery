package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"

	"github.com/slack-go/slack"

	"github.com/linode/linodego"
)

type tomlConfig struct {
	Slack slackConfig
}

type slackConfig struct {
	Channel string
	Token   string
}

var api *slack.Client

var config tomlConfig

var channel slack.Channel

func main() {
	// config
	if _, err := toml.DecodeFile("/etc/sink/sink.toml", &config); err != nil {
		log.Fatal(err)
	}

	api = slack.New(config.Slack.Token)

	// TODO: handle bad channel names
	channel = getSlackChannelByName(config.Slack.Channel)

	http.HandleFunc("/sink-slack", sinkSlackHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func sinkSlackHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var les []linodego.Event

	err := decoder.Decode(&les)
	if err != nil {
		log.Fatal(err)
	}

	for _, le := range les {
		message := fmt.Sprintf("%s %s %s %s\n", le.Entity.Type, le.Entity.Label, le.Action, le.Status)
		channelID, _, err := api.PostMessage(channel.ID, slack.MsgOptionText(message, false))
		if err != nil {
			// TODO: look into writing to stderr
			//fmt.Fprintln(os.Stderr, err)
			log.Println(err)
		} else {
			log.Printf("INFO {channel=%s} message successfully sent", channelID)
		}
	}
}

func getSlackChannelByName(name string) slack.Channel {
	// TODO: find channels more cleanly
	var sc slack.Channel

	params := slack.GetConversationsParameters{}
	channels, _, err := api.GetConversations(&params)
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range channels {
		if c.Name == name {
			sc = c
		}
	}

	return sc
}
