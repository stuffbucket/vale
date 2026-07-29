# 0008. LLM-assisted fix, shared model config, and progress UI

Status: Accepted

## Context

Reporting findings is only half of useful. Users want vale to *repair* a document,
not just point at errors. That needs an LLM, which vale already talks to for
`vale eval`. Three related needs: a rewrite mode, a place to configure the model
(endpoint, name, temperature) overridable per invocation, and clearer progress
for the long-running eval.

## Decision

1. **`internal/fix`** rewrites a document. It sends the original text plus the
   concise findings to a model and returns the corrected document (stripping a
   wrapping code fence). It reuses `internal/eval`'s client, so there is one HTTP
   path to the endpoint.
2. **`--fix` mode on `vale lint`.** `vale --fix <file>` prints the corrected
   document to stdout, or to `--output <file>`. It fixes STE findings and the
   opt-in slop markers together. Flags may be interspersed with the path
   (`parsePositional` re-parses around positionals, since Go's `flag` stops at the
   first non-flag).
3. **A shared `model` config** (`endpoint`, `name`, `temperature`). Temperature is
   a `*float64` so "unset" (server default) differs from an explicit 0. Every
   field is overridable on the command line, and temperature is also overridable
   through the MCP `fix_text` tool.
4. **`fix_text` MCP tool** exposes the same rewrite over MCP, with a `temperature`
   argument.
5. **charmbracelet for progress.** `vale eval` draws a live `bubbles/progress` bar
   to stderr when stderr is a terminal, via an `OnProgress` callback on
   `eval.Run`. This adds the charmbracelet dependency tree (pure Go, cgo-free).

## Consequences

- vale can repair documents, not only report on them, from the CLI and MCP.
- The dependency count grows (charmbracelet + yaml + goldmark). Still pure Go and
  `CGO_ENABLED=0`; cross-compilation is unaffected.
- The model config is one place, shared by eval and fix, overridable everywhere.
