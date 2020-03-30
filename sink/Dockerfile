FROM golang:1.14-buster
MAINTAINER Steven Jacobs <sjacobs@linode.com>

COPY . /app

WORKDIR /app

RUN go get -v -t -d ./... \
 && go build .

CMD ["/app/linode-event-sink"]
