VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test clean lint install

build:
	go build $(LDFLAGS) -o rctl ./cmd/rctl

test:
	go test ./... -count=1

test-v:
	go test ./... -count=1 -v

clean:
	rm -f rctl
	go clean ./...

lint:
	go vet ./...

install:
	VERSION=$(VERSION) ./scripts/install.sh
