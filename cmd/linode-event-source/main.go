package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

type tomlConfig struct {
	Source source
	Sink   sink
}

type source struct {
	URL   string
	Token string
	// TODO: handle time.Duration right
	Interval string
}

type sink struct {
	URL string
}

var config tomlConfig

type IngestService struct {
	Config tomlConfig
}

func forwardLinodeEvent(event linodego.Event, sink sink) {
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
	log.Printf("INFO {event=%d}: event forwarded successfully", event.ID)
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
func listLinodeEventsSince(linode linodego.Client, since time.Time) ([]linodego.Event, error) {
	opts := linodego.ListOptions{
		PageOptions: &linodego.PageOptions{Page: 1},
		PageSize:    25,
		Filter:      fmt.Sprintf(`{"created": {"+gte": "%s"}}`, since.Format("2006-01-02T15:04:05")),
	}

	events, err := linode.ListEvents(context.Background(), &opts)
	if err != nil {
		return []linodego.Event{}, fmt.Errorf("ERROR: failed to list events: %w", err)
	}

	return events, nil
}

func (service IngestService) Start(source source) {
	client := createLinodeClient(source)

	lastRun := time.Now().UTC()

	interval, err := time.ParseDuration(source.Interval)
	if err != nil {
		log.Fatal(err)
	}

	c := time.Tick(interval)

	for range c {
		go func() {
			log.Println("INFO: checking for new events")
			events, err := listLinodeEventsSince(client, lastRun)
			if err != nil {
				// TODO: look into writing to stderr
				//fmt.Fprintln(os.Stderr, err)
				log.Println(err)
			}

			for _, event := range events {
				forwardLinodeEvent(event, config.Sink)
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

	// service
	service := IngestService{
		Config: config,
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go service.Start(config.Source)
	waitGroup.Wait()
}
