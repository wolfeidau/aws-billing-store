APPNAME := aws-apilatency
STAGE ?= dev
BRANCH ?= master

GOLANGCI_VERSION = v1.46.2

GIT_HASH := $(shell git rev-parse --short HEAD)

.PHONY: ci
ci: test build

.PHONY: lint
lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run -v

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: build
build:
	CGO_ENABLED=0 GOAMD64=v2 go build -ldflags "-s -w -X main.commit=$(GIT_HASH)" -o dist/ ./cmd/...

.PHONY: archive
archive:
	@echo "--- build an archive"
	@cd dist && zip -X -9 ./handler.zip *-lambda
