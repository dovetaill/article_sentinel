CONFIG ?= configs/config.example.yaml

.PHONY: run test verify build

run:
	go run ./cmd/server -config $(CONFIG)

test:
	go test ./...

verify:
	go test ./...

build:
	go build ./cmd/server
