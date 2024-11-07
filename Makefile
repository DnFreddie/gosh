.PHONY: build test clean install

build:
	go build

test:
	go test -v ./...

clean:
	go mod tidy

install:
	go install ./...

