.PHONY: build build-all test vet lint install clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BINARY  := jenkins
LDFLAGS := -s -w

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

build-all:
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/jenkins-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/jenkins-darwin-amd64 .
	GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/jenkins-linux-arm64  .
	GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/jenkins-linux-amd64  .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

install: build
	go install -trimpath -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY)
	rm -rf dist/
