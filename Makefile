.PHONY: build test clean install

build:
	go build -o g

test:
	go test  ./...

clean:
	go mod tidy

install:
	go install ./...

