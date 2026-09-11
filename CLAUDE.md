# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`viber` is a scaffold ("vibe coding project kickstarter") intended for a Go CLI. There is currently no source code and no `go.mod`; the Go-flavored `.gitignore` is the only clue to the intended stack. If asked to build features, scaffold `go mod init` and a `main.go` (or `cmd/`) before implementing.

## Conventions

- Format with `goimports -w` (fallback: `gofmt -w`). A `PostToolUse` hook auto-formats `.go` files after every Write/Edit — don't format by hand.
- A `Stop` hook runs `golangci-lint run ./... && go test ./...` at the end of each turn (no-op until `go.mod` exists). Treat a red result as blocking. Lint config: `.golangci.yml` (v2, `default: standard`).

## Skills

- `/commit-push-pr` — user-invoked: stage, commit, push, open a PR via `gh`.
