# AGENTS.md

This file gives instructions for coding agents. It uses the same content as
`CLAUDE.md`. The linter in this repository checks this file. Keep it green.

## Project summary

Vale is one Go binary with two faces. The first face is a command-line linter
for Simplified Technical English. The second face is an MCP server over stdio.
Both faces use the same engine, so they give the same findings.

Vale builds part of itself. The `vale gen` command reads the vendored OpenSTE
wordset. It writes the vocabulary rules from that wordset. The continuous
integration job lints this file with the new binary. The job fails on a new
error.

## Rules for agents

- Write all prose in short sentences.
- Do not use contractions. Write the full words.
- Keep each list item to 20 words or less.
- Keep each sentence in a paragraph to 25 words or less.
- Do not edit the generated file `internal/vocab/substitutions_gen.go` by hand.
- Run `vale gen` to make that file again after a change to the wordset.
- Add one rule for each file. Add a table-driven test for each rule.
- Give each rule a stable identifier, a severity, and a fix hint.

## How to add a rule

1. Add a file in `internal/rules/`.
2. Make a type that satisfies the `lint.Rule` interface.
3. Register the rule in `internal/rules/rules.go`.
4. Add a table-driven test with good cases and bad cases.
5. Run `go test -race ./...`.

## How to verify your work

Run these commands before you commit:

- `go build ./...` and `go vet ./...`
- `go test -race ./...`
- `golangci-lint run`
- `go run ./cmd/vale lint CLAUDE.md AGENTS.md`

The last command must pass with no error.
