package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

// LogEvent is
// i.e.
type LogEvent struct {
	Host      string `json:"host"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	// TODO: additional fields. something like logrus.
	// type Fields map[string]interface{}
}

// EventMetric is
// i.e.
// type MetricEvent struct {}

func isNew(event linodego.Event) bool {
	return true
}

func ListNewLogEvents(linode linodego.Client) ([]linodego.Event, error) {
	filter := fmt.Sprintf("{}")
	events, err := linode.ListEvents(context.Background(), &linodego.ListOptions{Filter: filter})

	if err != nil {
		log.Fatal("Error getting Events, expected struct, got error %v", err)
	}

	return events, err
}

func MarshalLogEvent(event linodego.Event) {
	switch event.Entity.Type {
	case "community_like":
		fmt.Printf("info: skipping event. id=%d action=%s type=%s", event.ID, event.Action, event.Entity.Type)
	default:
		fmt.Printf("entity: %v\nevent: %v\n", event.Entity.Type, event)
	}
}

func ForwardLogEvent() {

}

func main() {
	apiKey, ok := os.LookupEnv("LINODE_TOKEN")
	if !ok {
		log.Fatal("Could not find LINODE_TOKEN, please assert it is set.")
	}
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient := linodego.NewClient(oauth2Client)
	//linodeClient.SetDebug(true)

	events, err := ListNewLogEvents(linodeClient)
	if err != nil {
		log.Fatal(err)
	}

	for _, event := range events {
		MarshalLogEvent(event)
	}
}
