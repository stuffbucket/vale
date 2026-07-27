# CLAUDE.md

This file gives instructions for coding agents that work in this repository.
The rules use Simplified Technical English. The linter in this repository
checks this file. Keep it green.

## What this project is

Vale is one Go binary. It has two faces. The first face is a command-line
linter for Simplified Technical English. The second face is an MCP server. The
server gives the same engine to an agent through stdio.

The binary also builds itself. The `vale gen` command reads the vendored
OpenSTE wordset. It writes the vocabulary rules from that wordset. The
continuous integration job lints this file and `AGENTS.md`. The job fails on a
new error.

## Rules for agents

- Write all prose in short sentences.
- Do not use contractions. Write the full words.
- Keep each list item to 20 words or less.
- Keep each sentence in a paragraph to 25 words or less.
- Do not edit the generated file `internal/vocab/substitutions_gen.go` by hand.
- Run `vale gen` to make that file again after a change to the wordset.
- Keep the OpenSTE license and the attribution.

## Layout

- `cmd/vale/` holds the command-line tool.
- `internal/lint/` holds the engine, the tokenizer, and the rule interface.
- `internal/rules/` holds one file for each rule.
- `internal/vocab/` holds the generated vocabulary and the generator.
- `internal/mcp/` holds the stdio MCP server.
- `internal/config/` holds the configuration loader.
- `skills/` holds the Claude Code skill.
- `docs/decisions/` holds the architecture decision records.

## How to verify your work

Run these commands before you commit:

- `go build ./...` and `go vet ./...`
- `go test -race ./...`
- `golangci-lint run`
- `go run ./cmd/vale gen`, then make sure that git shows no difference.
- `go run ./cmd/vale lint CLAUDE.md AGENTS.md`

The last command must pass with no error.
