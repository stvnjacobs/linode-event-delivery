package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/slack-go/slack"

	"github.com/linode/linodego"
	//"golang.org/x/oauth2"
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

// LinodeEvent represents a linodego.Event with additional metadata
type LinodeEvent struct {
	Source    string         `json:"source"`
	Event     linodego.Event `json:"event"`
	Timestamp time.Time      `json:"timestamp"`
}

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

	var les []LinodeEvent

	err := decoder.Decode(&les)
	if err != nil {
		log.Fatal(err)
	}

	for _, le := range les {
		message := fmt.Sprintf("%s - %s %s %s %s\n", le.Source, le.Event.Entity.Type, le.Event.Entity.Label, le.Event.Action, le.Event.Status)
		channelID, _, err := api.PostMessage(channel.ID, slack.MsgOptionText(message, false))
		if err != nil {
			log.Fatal(err)
		}
		log.Print(fmt.Sprintf("INFO {channel=%s} message successfully sent", channelID))
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

	for _, c :=range channels {
		if c.Name == name {
			sc = c
		}
	}

	return sc
}

//buf := new(bytes.Buffer)
//buf.ReadFrom(r.Body)
//body := buf.String()

//log.Print(json.RawMessage(body))

		//err := json.Unmarshal([]byte(body), &le)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	log.Fatal(err)
		//}
		//log.Print(le)

		//eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: "TOKEN"}))
		//if e != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//}

		//if eventsAPIEvent.Type == slackevents.URLVerification {
		//	var r *slackevents.ChallengeResponse
		//	err := json.Unmarshal([]byte(body), &r)
		//	if err != nil {
		//		w.WriteHeader(http.StatusInternalServerError)
		//	}
		//	w.Header().Set("Content-Type", "text")
		//	w.Write([]byte(r.Challenge))
		//}
		//if eventsAPIEvent.Type == slackevents.CallbackEvent {
		//	innerEvent := eventsAPIEvent.InnerEvent
		//	switch ev := innerEvent.Data.(type) {
		//	case *slackevents.AppMentionEvent:
		//		api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
		//	}
		//}
//	})
//}
