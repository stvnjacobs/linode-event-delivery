package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"

	badger "github.com/dgraph-io/badger/v2"
)

type tomlConfig struct {
	DB     database
	Source source
	Sink   sink
}

type database struct {
	Path string
}

type source struct {
	URL   string
	Token string
	// TODO: handle time.Duration right
	Interval string
}

type sink struct {
	Type string
	// TODO: handle url right.
	URL string
}

var config tomlConfig

var db *badger.DB

type IngestService struct {
	DB     *badger.DB
	Config tomlConfig
}

// LinodeEvent represents a linodego.Event with additional metadata
type LinodeEvent struct {
	Source    string         `json:"source"`
	Event     linodego.Event `json:"event"`
	Timestamp time.Time      `json:"timestamp"`
}

// PopulateLinodeEvent is responsible for taking a linodego.Event and adding additional metadata
func populateLinodeEvent(event linodego.Event) LinodeEvent {
	return LinodeEvent{
		Event:     event,
		Timestamp: time.Now(),
	}
}

func filterNewLinodeEvents(db *badger.DB, events []linodego.Event) []linodego.Event {
	var newEvents []linodego.Event

	for _, event := range events {
		if isEventNew(db, event) {
			newEvents = append(newEvents, event)
		}
	}

	return newEvents
}

func isEventNew(db *badger.DB, event linodego.Event) bool {
	// TODO: make prefix configurable
	prefix := []byte(fmt.Sprintf("linode-account-event-%s-", event.Status))
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

// TODO: stop passing around just the db
func markLinodeEventAsSent(db *badger.DB, event linodego.Event) {
	// TODO: make prefix configurable
	prefix := []byte(fmt.Sprintf("linode-account-event-%s-", event.Status))

	err := db.Update(func(txn *badger.Txn) error {
		_ = txn.Set([]byte(fmt.Sprintf("%s-%d", prefix, event.ID)), []byte(strconv.Itoa(1)))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func forwardLinodeEvent(event LinodeEvent, sink sink) {
	conn, err := net.Dial("tcp", sink.URL)
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
	log.Print(fmt.Sprintf("INFO {event=%d}: event forwarded successfully", event.Event.ID))
}

func createLinodeClient(config source) linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	client := linodego.NewClient(oauth2Client)
	//client.SetDebug(true)

	return client
}

// TODO: handle more than 25 events
func listLinodeEventsSince(db *badger.DB, linode linodego.Client, since time.Time) []linodego.Event {
	opts := linodego.ListOptions{
		PageOptions: &linodego.PageOptions{Page: 1},
		PageSize:    25,
		Filter:      fmt.Sprintf(`{"created": {"+gte": "%s"}}`, since.Format("2006-01-02T15:04:05")),
	}

	allEvents, err := linode.ListEvents(context.Background(), &opts)
	if err != nil {
		log.Fatal("Error getting Events, expected struct, got error %v", err)
	}

	filteredEvents := filterNewLinodeEvents(db, allEvents)

	return filteredEvents
}

func (service IngestService) Start(source source) {
	client := createLinodeClient(source)

	lastRun := time.Now()

	interval, err := time.ParseDuration(source.Interval)
	if err != nil {
		log.Fatal(err)
	}

	c := time.Tick(interval)

	for range c {
		go func() {
			log.Print(fmt.Sprintf("INFO: checking for new events"))
			events := listLinodeEventsSince(db, client, lastRun)

			for _, event := range events {
				// add extra info
				// TODO: fix odd type change
				e := populateLinodeEvent(event)
				// send it
				forwardLinodeEvent(e, config.Sink)
				// mark it as sent
				markLinodeEventAsSent(db, event)
			}

			lastRun = lastRun.Add(interval)
		}()
	}
}

func main() {
	// config
	if _, err := toml.DecodeFile("/etc/source/source.toml", &config); err != nil {
		log.Fatal(err)
	}

	// persistence
	// TODO: learn how to do errors right
	db, _ = badger.Open(badger.DefaultOptions(config.DB.Path))

	// service
	service := IngestService{
		DB:     db,
		Config: config,
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go service.Start(config.Source)
	waitGroup.Wait()
}
