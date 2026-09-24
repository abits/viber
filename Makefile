SHELL := /usr/bin/env bash

VERSION := $(shell cat VERSION)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

.PHONY: all build install test lint fmt tidy man completions release \
        dogfood dogfood-check lint-md repo-init sync-issues \
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

# viber is scaffolded with itself. DOGFOOD_FILES are owned by the embedded
# template set: `make dogfood` copies them verbatim from a fresh render, and
# `make dogfood-check` (run in CI) fails when any of them has drifted. Edit
# them under internal/templates/default/, never in place. CLAUDE.md,
# README.md, Makefile, .gitignore, and .claude/settings.json are repo-owned
# supersets of their templates and are merged by hand.
DOGFOOD_NAME := viber
DOGFOOD_DESC := Scaffold a vibe-coding project wired for Claude Code + OpenSpec.
DOGFOOD_FILES := AGENTS.md .markdownlint.yaml \
	scripts/repo-init.sh scripts/sync-issues.sh \
	.claude/commands/repo-init.md .claude/commands/sync-issues.md \
	$(patsubst internal/templates/default/%.tmpl,%,$(wildcard \
		internal/templates/default/.claude/agents/*.tmpl \
		internal/templates/default/.claude/rules/*.tmpl))

dogfood:
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	go run ./tools/render-template -name "$(DOGFOOD_NAME)" -desc "$(DOGFOOD_DESC)" "$$tmp/out" && \
	for f in $(DOGFOOD_FILES); do \
		mkdir -p "$$(dirname "$$f")" && cp -f "$$tmp/out/$$f" "$$f" || exit 1; \
	done && \
	echo "dogfood: wrote $(words $(DOGFOOD_FILES)) template-owned files"

dogfood-check:
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	go run ./tools/render-template -name "$(DOGFOOD_NAME)" -desc "$(DOGFOOD_DESC)" "$$tmp/out" && \
	rc=0 && for f in $(DOGFOOD_FILES); do \
		diff -u "$$tmp/out/$$f" "$$f" || rc=1; \
	done && \
	if [ $$rc -ne 0 ]; then echo "dogfood-check: template-owned files drifted; run 'make dogfood'" >&2; fi && \
	exit $$rc

# Scaffold targets, identical to the ones `viber init` generates.
lint-md:
	markdownlint-cli2 "**/*.md"

repo-init:
	bash scripts/repo-init.sh

sync-issues:
	bash scripts/sync-issues.sh

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
