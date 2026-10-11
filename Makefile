BINARY := entropy
PKG := ./cmd/entropy
GO ?= go
DEMO_HOME ?= $(CURDIR)/.demo

.PHONY: build test test-race lint fmt vet tidy demo clean run

## build: compile the CLI into ./bin/entropy
build:
	$(GO) build -o bin/$(BINARY) $(PKG)

## install: install the CLI into GOPATH/bin
install:
	$(GO) install $(PKG)

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

## lint: run gofmt check and vet
lint: fmt vet

## tidy: sync go.mod/go.sum
tidy:
	$(GO) mod tidy

## demo: run the end-to-end scenario in a throwaway data dir
demo:
	./scripts/demo.sh

## run: build and run the CLI (ARGS="..." make run)
run:
	$(GO) run $(PKG) $(ARGS)

## clean: remove build output and demo data
clean:
	rm -rf bin $(DEMO_HOME)
