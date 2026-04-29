.PHONY: run test lint build

run:
	go run .

test:
	go test ./...
	go test -race ./...

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...

build:
	go build -o bin/tg_bot_golang .
