# linode-event-delivery

## Components

### `linode-event-source`

Pulls events and forwards them to configured sink.

#### Configuration

``` toml
# file: /etc/source/source.toml

[source]
url = "https://api.linode.com/v4"
token = "CHANGEME" # must have at least "accounts:read_only" and "events:read_only"
interval = "10s"   # format follows https://golang.org/pkg/time/#ParseDuration

[sink]
url = "localhost:9000"
```

### `linode-event-sink-slack`

Handles incoming events, forwarding them to configured Slack channel.

#### Configuration

``` toml
# file: /etc/sink/sink.toml

[slack]
token = "xoxb-example-token"
channel = "notification-linode"
```

## Examples

The repository provides an example docker-compose file, showing how to put a tool like [Vector](https://vector.dev) between the source and the sink. Using this topology enables multiple account sources writing to multiple sinks, Slack just being just one of them. To showcase this, the docker-compose file also writes events to an Elasticsearch database, but this could also additionally write to S3, Kafka, or any of the other [sinks that Vector supports](https://vector.dev/docs/reference/configuration/sinks/)

### Configuration

To get started copy the example configs and edit.

``` sh
cp -r ./example ./config
```

## Usage

``` sh
docker-compose up -d
```
