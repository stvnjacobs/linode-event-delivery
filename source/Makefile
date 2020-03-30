.PHONY: reset clean install build

# TODO: confirmation prompt
reset:
	rm -r /tmp/badger

clean:
	rm linode-event-delivery

install:
	go get -v -t -d ./...

# TODO: learn makefiles. we're using compilers now.
# TODO: target binary
build: install
	go build .

run: install
	go run .
