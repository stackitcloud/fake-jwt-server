SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)
export PATH := $(PWD)/bin:$(PATH)
export CGO_ENABLED = 0

# https://github.com/golangci/golangci-lint/releases
GOLANGCI_VERSION = 1.56.2

VERSION ?= 0.0.1
IMAGE_TAG_BASE ?= stackitcloud/fake-jwt-server
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

download:
	go mod download

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o ./bin/fake-jwt-server -v cmd/fakejwtserver/main.go

.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o ./bin/fake-jwt-server -v cmd/fakejwtserver/main.go
	docker build -t $(IMG) -f Dockerfile .
	rm fake-jwt-server

GOLANGCI_LINT = bin/golangci-lint-$(GOLANGCI_VERSION)
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b bin v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint "$(@)"

lint: $(GOLANGCI_LINT) download
	$(GOLANGCI_LINT) run -v

run:
	go run cmd/fakejwtserver/main.go
