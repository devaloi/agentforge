.PHONY: build test lint run clean fmt vet

BINARY := agentforge
PKG := ./...

build:
	go build -o bin/$(BINARY) ./cmd/agentforge/

test:
	go test -race -count=1 $(PKG)

lint:
	golangci-lint run

fmt:
	gofmt -w .

vet:
	go vet $(PKG)

run:
	go run ./cmd/agentforge/ run "Build a REST API for a todo app in Go"

clean:
	rm -rf bin/ output/

all: fmt vet lint test build
