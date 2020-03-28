FROM golang:1.14-buster
MAINTAINER Steven Jacobs <sjacobs@linode.com>

RUN apt-get -yq update && \
    apt-get -yq install make

COPY . /app

WORKDIR /app

RUN make build

CMD ["make", "run"]
