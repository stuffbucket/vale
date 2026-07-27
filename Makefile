GO           ?= go
BINARY       := vale
PKG          := ./cmd/vale
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS      := -s -w -X main.version=$(VERSION)
COVERPROFILE := coverage.txt

export CGO_ENABLED := 0

.PHONY: all build test cover lint fuzz gen mutation self-lint snapshot check clean

all: build

## build: compile the vale binary.
build:
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) $(PKG)

## test: run the full test suite with the race detector.
test:
	$(GO) test -race ./...

## cover: write a coverage profile and print the per-func summary + total.
cover:
	$(GO) test -race -covermode=atomic -coverprofile=$(COVERPROFILE) ./...
	$(GO) tool cover -func=$(COVERPROFILE) | tail -n 1

## lint: run golangci-lint.
lint:
	golangci-lint run

## fuzz: short fuzz smoke over the lint fuzz targets.
fuzz:
	$(GO) test -run=^$$ -fuzz=FuzzParse -fuzztime=20s ./internal/lint
	$(GO) test -run=^$$ -fuzz=FuzzEngineRun -fuzztime=20s ./internal/lint

## gen: regenerate the vocabulary rules from the wordset.
gen:
	$(GO) generate ./...

## mutation: run gremlins mutation testing (uses .gremlins.yaml).
mutation:
	gremlins unleash

## self-lint: lint the repo docs as Simplified Technical English.
self-lint: build
	./$(BINARY) lint CLAUDE.md AGENTS.md

## snapshot: build a local goreleaser snapshot for all targets.
snapshot:
	goreleaser build --snapshot --clean

## check: build, vet, lint, and test.
check: build
	$(GO) vet ./...
	golangci-lint run
	$(GO) test -race ./...

## clean: remove build and coverage output.
clean:
	rm -f $(BINARY) $(COVERPROFILE) coverage.html
	rm -rf dist
