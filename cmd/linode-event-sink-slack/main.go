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

	_channel, err := getSlackChannelByName(config.Slack.Channel)
	if err != nil {
		log.Fatal(err)
	}

	channel = _channel
	log.Printf("INFO {channel=%s} channel found", channel.ID)

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

func getSlackChannelByName(name string) (slack.Channel, error) {
	// TODO: find channels more cleanly
	params := slack.GetConversationsParameters{ExcludeArchived: true}
	cur := "start"
	for cur != "" {
		channels, cur, err := api.GetConversations(&params)
		if err != nil {
			return slack.Channel{}, err
		}
		for _, c := range channels {
			if c.Name == name {
				return c, nil
			}
		}
		params.Cursor = cur
	}

	return slack.Channel{}, fmt.Errorf("no matching channel found")
}
