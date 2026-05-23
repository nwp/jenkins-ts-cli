.PHONY: build test lint install clean

BINARY := jenkins
GOFLAGS := -trimpath

build:
	go build $(GOFLAGS) -o $(BINARY) .

install:
	go install $(GOFLAGS) .

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
