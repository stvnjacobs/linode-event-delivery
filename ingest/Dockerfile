FROM golang:1.14
MAINTAINER Steven Jacobs <sjacobs@linode.com>

COPY . /app

WORKDIR /app

RUN go get -v -t -d ./... && \
    go build

CMD ["go", "run", "."]
