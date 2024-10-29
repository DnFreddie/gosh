.PHONY: build test install clean

build:
    go build

test:
    go test -v ./...

clean:
    go mod tidy

install:
    go get

