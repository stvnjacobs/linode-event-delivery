# linode-event-delivery

## Requirements

`LINODE_TOKEN` with `read` scope on `Account` and `Events`

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
