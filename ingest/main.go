package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"

	badger "github.com/dgraph-io/badger/v2"
)

const (
	refreshIntervalSec = 10
)

var db *badger.DB

// TODO: investigate other shutdown handlers.
var shutdownCh chan struct{}

// LogEvent is a mapping of a Vector Log Event
// https://vector.dev/docs/about/data-model/log/
type LogEvent struct {
	Host      string `json:"host"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	// TODO: additional fields. something like logrus.
	// type Fields map[string]interface{}
}

// MetricEvent is a mapping of a Vector Metric Event
// https://vector.dev/docs/about/data-model/event/
// type MetricEvent struct {}

// store lowest event.ID which all lower event.IDs are 100% completed
// find the page of the eventID and only query those pages
// send along changes
func ListNewLogEvents(db *badger.DB, linode linodego.Client) []linodego.Event {
	filter := fmt.Sprintf("{}")
	opts := linodego.NewListOptions(1, filter)

	allEvents, err := linode.ListEvents(context.Background(), opts)
	if err != nil {
		log.Fatal("Error getting Events, expected struct, got error %v", err)
	}

	filteredEvents := FilterNewLogEvents(db, allEvents)

	return filteredEvents
}

func FilterNewLogEvents(db *badger.DB, events []linodego.Event) []linodego.Event {
	var newEvents []linodego.Event

	for _, event := range events {
		if isEventNew(db, event) {
			fmt.Println("I seent a new event!")
			newEvents = append(newEvents, event)
		} else {
			fmt.Println("I already seent that event. Boring!")
		}
	}

	return newEvents
}

func isEventNew(db *badger.DB, event linodego.Event) bool {
	// TODO: make prefix configurable
	prefix := []byte("linode-account-event")
	isNew := false

	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(fmt.Sprintf("%s-%d", prefix, event.ID)))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				log.Fatal(err)
			}
			isNew = true
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return isNew
}

// TODO: update event as Sent not Seen
// TODO: stop passing around just the db
func MarkNewLogEventAsSeen(db *badger.DB, event linodego.Event) {
	// TODO: make prefix configurable
	prefix := []byte("linode-account-event")

	err := db.Update(func(txn *badger.Txn) error {
		_ = txn.Set([]byte(fmt.Sprintf("%s-%d", prefix, event.ID)), []byte(strconv.Itoa(1)))
		fmt.Sprintf("%s-%d", prefix, event.ID)

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func MarshalLogEvent(event linodego.Event) LogEvent {
	message, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}

	return LogEvent{
		Host:    "foo",
		Message: fmt.Sprintf("%s", message),
		// TODO: fix timestamps for created _and_ updated
		Timestamp: event.Created.Format(time.RFC3339),
	}
}

// TODO: FilterLogEvent
//	switch event.Entity.Type {
//	case "community_like":
//		fmt.Printf("info: skipping event. id=%d action=%s type=%s", event.ID, event.Action, event.Entity.Type)
//	default:
//		fmt.Printf("entity: %v\nevent: %v\n", event.Entity.Type, event)
//	}

func ForwardLogEvent(event LogEvent) {
	conn, err := net.Dial("tcp", "vector:9000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// TODO: learn how to do errors right
	message, merr := json.Marshal(event)
	if merr != nil {
		log.Fatal(merr)
	}

	// send to socket
	fmt.Fprintf(conn, string(message)+"\n")
}

func init() {
	// Create channel to listen to OS interrupt signals
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)
	go onExit()

	// TODO: update path to /var/lib/ingest/data/badger
	// TODO: learn how to do errors right
	db, _ = badger.Open(badger.DefaultOptions("/tmp/badger"))
}

func onExit() {
	<-shutdownCh
	db.Close()
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

	events := ListNewLogEvents(db, linodeClient)

	for _, event := range events {
		ForwardLogEvent(MarshalLogEvent(event))
		MarkNewLogEventAsSeen(db, event)
	}
}
