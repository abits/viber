SHELL := /usr/bin/env bash

VERSION := $(shell cat VERSION)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

.PHONY: all build install test lint fmt tidy man completions release \
        bump-patch bump-minor bump-major _bump

all: build

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/viber ./cmd/viber

install: build
	@mkdir -p $$HOME/bin
	install -m 0755 bin/viber $$HOME/bin/viber
	@echo "installed $$HOME/bin/viber"

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	goimports -w .

tidy:
	go mod tidy

man: build
	@mkdir -p man
	./bin/viber gen-man man

completions: build
	@mkdir -p completions
	./bin/viber completion bash > completions/viber.bash
	./bin/viber completion zsh  > completions/viber.zsh
	./bin/viber completion fish > completions/viber.fish

release:
	goreleaser release --clean

bump-patch:
	@$(MAKE) --no-print-directory _bump PART=patch
bump-minor:
	@$(MAKE) --no-print-directory _bump PART=minor
bump-major:
	@$(MAKE) --no-print-directory _bump PART=major

_bump:
	@git diff --quiet && git diff --cached --quiet || { echo "working tree not clean; commit or stash first"; exit 1; }
	@new=$$(awk -F. -v p=$(PART) '{if(p=="patch")$$3++;else if(p=="minor"){$$2++;$$3=0}else{$$1++;$$2=0;$$3=0}print $$1"."$$2"."$$3}' VERSION); \
	echo "$$new" > VERSION; \
	git add VERSION; \
	git commit -m "chore: bump to v$$new"; \
	git tag "v$$new"; \
	echo "bumped to v$$new"
