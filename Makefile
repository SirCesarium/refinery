BINARY_NAME=refinery
VERSION=$$(git describe --tags --always --dirty)
COMMIT=$$(git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: all build test clean help smoke-test

all: test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/refinery

smoke-test: ## Run smoke tests inside a Docker container
	docker build -t refinery-smoke-test -f tests/smoke/Dockerfile .
	docker run --rm refinery-smoke-test

smoke-test-local: build ## Run smoke tests locally
	bash ./scripts/smoke-test.sh bin/$(BINARY_NAME)

run:
	go run ./cmd/refinery

test:
	go test -v ./...

fmt:
	go fmt ./...
	goimports -w .

clean:
	rm -rf bin/
	go clean

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'