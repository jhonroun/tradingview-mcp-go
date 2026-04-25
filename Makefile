# tradingview-mcp-go — build targets
MODULE  := github.com/rusernam/tradingview-mcp-go
LDFLAGS := -ldflags="-s -w"
BIN     := bin

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

EXT := $(if $(filter windows,$(GOOS)),.exe,)

.PHONY: all build build-all install test test-verbose clean release package

all: build

## build — compile for the current platform into bin/
build:
	@mkdir -p $(BIN)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BIN)/tvmcp$(EXT) ./cmd/tvmcp
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BIN)/tv$(EXT)    ./cmd/tv
	@echo "Built: $(BIN)/tvmcp$(EXT)  $(BIN)/tv$(EXT)"

## build-all — cross-compile for Windows / Linux / macOS (amd64 + arm64)
build-all:
	@for os in windows linux darwin; do \
		for arch in amd64 arm64; do \
			ext=$$([ "$$os" = "windows" ] && echo ".exe" || echo ""); \
			dir=$(BIN)/$$os-$$arch; \
			mkdir -p $$dir; \
			echo "  $$os/$$arch → $$dir"; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$dir/tvmcp$$ext ./cmd/tvmcp; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$dir/tv$$ext    ./cmd/tv; \
		done; \
	done
	@echo "All builds in $(BIN)/"

## install — install both binaries into GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/tvmcp ./cmd/tv
	@echo "Installed to $$(go env GOPATH)/bin/"

## test — run all unit tests
test:
	go test ./...

## test-verbose
test-verbose:
	go test -v ./...

## clean — remove bin/
clean:
	rm -rf $(BIN)

## package — full release bundles: binaries + agents + skills + prompts + scripts + installers
package:
	bash scripts/package.sh

## release — alias for package (backwards compat)
release: package
