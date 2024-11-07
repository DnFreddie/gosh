.PHONY: build test clean install

build:
	go build -o g

test:
	go test -v ./...

clean:
	go mod tidy

install:
	go install ./...

