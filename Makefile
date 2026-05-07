CONFIG ?= configs/config.example.yaml

.PHONY: run test verify build

run:
	go run ./cmd/server -config $(CONFIG)

test:
	go test ./...

verify:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server
