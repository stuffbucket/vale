GO           ?= go
BINARY       := vale
PKG          := ./cmd/vale
BI           := github.com/stuffbucket/vale/internal/buildinfo
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT       ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
BRANCH       ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)
DATE         ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS      := -s -w \
	-X $(BI).version=$(VERSION) \
	-X $(BI).commit=$(COMMIT) \
	-X $(BI).branch=$(BRANCH) \
	-X $(BI).date=$(DATE)
COVERPROFILE := coverage.txt

export CGO_ENABLED := 0

.PHONY: all build test cover lint shellcheck fuzz gen mutation self-lint snapshot check setup clean

all: build

## setup: install developer tools (idempotent). Installs shellcheck when it is
## not already on PATH, via Homebrew or apt.
setup:
	@if command -v shellcheck >/dev/null 2>&1; then \
		echo "shellcheck present: $$(shellcheck --version | awk '/version:/{print $$2}')"; \
	elif command -v brew >/dev/null 2>&1; then \
		echo "installing shellcheck via Homebrew..."; brew install shellcheck; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "installing shellcheck via apt..."; sudo apt-get update && sudo apt-get install -y shellcheck; \
	else \
		echo "install shellcheck manually: https://github.com/koalaman/shellcheck#installing"; exit 1; \
	fi

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

## shellcheck: check shell scripts as POSIX sh (run `make setup` to install it).
shellcheck:
	@files=$$(find . -name '*.sh' -not -path './.git/*' -not -path './dist/*'); \
	if [ -n "$$files" ]; then shellcheck -s sh $$files && echo "shellcheck: clean"; \
	else echo "shellcheck: no shell scripts"; fi

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

## check: build, vet, lint, shellcheck, and test.
check: build shellcheck
	$(GO) vet ./...
	golangci-lint run
	$(GO) test -race ./...

## clean: remove build and coverage output.
clean:
	rm -f $(BINARY) $(COVERPROFILE) coverage.html
	rm -rf dist
