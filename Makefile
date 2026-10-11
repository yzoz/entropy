BINARY  := entropy
PKG     := ./cmd/entropy
GO      ?= go
DEMO_HOME ?= $(CURDIR)/.demo

# Version metadata embedded into the binary (see internal/cli/version.go).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/yzoz/entropy/internal/cli.version=$(VERSION) \
	-X github.com/yzoz/entropy/internal/cli.commit=$(COMMIT) \
	-X github.com/yzoz/entropy/internal/cli.date=$(DATE)

.PHONY: build install test test-race lint fmt vet tidy demo run snapshot clean

## build: compile the CLI into ./bin/entropy with version metadata
build:
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

## install: install the CLI into GOPATH/bin
install:
	$(GO) install -ldflags "$(LDFLAGS)" $(PKG)

## test: run the full test suite
test:
	$(GO) test ./...

## test-race: run the test suite under the race detector
test-race:
	$(GO) test -race ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## fmt: format all Go sources
fmt:
	gofmt -w ./cmd ./internal

## lint: run gofmt and vet
lint: fmt vet

## tidy: sync go.mod/go.sum
tidy:
	$(GO) mod tidy

## demo: run the end-to-end scenario in a throwaway data dir
demo:
	./scripts/demo.sh

## run: build and run the CLI (ARGS="..." make run)
run:
	$(GO) run -ldflags "$(LDFLAGS)" $(PKG) $(ARGS)

## snapshot: build release artifacts locally with GoReleaser (no publish)
snapshot:
	goreleaser release --snapshot --clean

## clean: remove build output and demo data
clean:
	rm -rf bin dist $(DEMO_HOME)
