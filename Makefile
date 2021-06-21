# https://github.com/hashicorp/terraform-provider-scaffolding/blob/main/.github/workflows/test.yml

BINARY_NAME=terraform-provider-sealedsecrets
BUILD_PATH=build
VERSION?=$(shell git describe --tags --abbrev=0)

GO_CMD=go

all: clean deps fmt build

build: 
	$(GO_CMD) build -o $(BINARY_NAME)@${VERSION} -v

build_darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO_CMD) build -o $(BINARY_NAME)_darwin_amd64_${VERSION} -v

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_CMD) build -o $(BINARY_NAME)_linux_amd64_${VERSION} -v

clean: 
	$(GO_CMD) clean
	rm -rf $(BINARY_NAME)*

deps:
	$(GO_CMD) mod download

fmt:
	$(GO_CMD) fmt ./...

tidy:
	$(GO_CMD) mod tidy
