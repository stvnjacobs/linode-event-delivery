# linode-event-delivery

## Components

### source

Pulls events, deduplicates, adds additional metadata and POSTs to configured endpoint.

### vector

https://vector.dev/docs/about/what-is-vector/

sinks:
- http
- elasticsearch
- kafka?

### sink

Vector POSTs HTTP event, sink forwards it to configured Slack channel.

## Configuration

`LINODE_TOKEN` _(default: `null`)_ **REQUIRED**: Must have with `read` scope on `Account` and `Events`

`LED_CONFIG_DIR` _(default: `./config`)_: Path to configuration e.g. /etc/linode-event-delivery

To get started copy the example configs and edit.

``` sh
cp -r ./example ./config
```

## Usage

``` sh
docker-compose up -d
```
