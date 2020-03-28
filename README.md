# linode-event-delivery

## Requirements

`LINODE_TOKEN` with `read` scope on `Account` and `Events`

## Components

### ingest

```
linodego events -> vector source
```

### vector

https://vector.dev/docs/about/what-is-vector/

sinks:
- http
- elasticsearch
- kafka?

### delivery

```
vector http sink -> slack
```

can we send it to a websocket?
