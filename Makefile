##
# linode-event-delivery
#
# @file
# @version 0.1

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linode-event-sink-slack ./cmd/linode-event-sink-slack
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linode-event-source ./cmd/linode-event-source

.PHONY: build-docker
build-docker:
	docker build -f build/package/Dockerfile-linode-event-source -t linode-event-source .
	docker build -f build/package/Dockerfile-linode-event-sink-slack -t linode-event-sink-slack .

.PHONY: clean
clean:
	rm -rf ./dist
# end
