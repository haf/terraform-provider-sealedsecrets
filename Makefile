# https://github.com/hashicorp/terraform-provider-scaffolding/blob/main/.github/workflows/test.yml

BINARY_NAME=terraform-provider-sealedsecrets
BUILD_PATH=build
VERSION?=0.2.4

GO_CMD=go

all: clean deps build

build: 
	$(GO_CMD) build -o $(BINARY_NAME)@v${VERSION} -v

build_darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO_CMD) build -o $(BINARY_NAME)_darwin_amd64_v${VERSION} -v

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_CMD) build -o $(BINARY_NAME)_linux_amd64_v${VERSION} -v

clean: 
	$(GO_CMD) clean
	rm -rf $(BINARY_NAME)*

deps:
	$(GO_CMD) mod download

fmt:
	$(GO_CMD) fmt

tidy:
	$(GO_CMD) mod tidy