# 1. A single, pure-Go binary with two faces

- Status: Accepted

## Context

Vale must be a Simplified Technical English linter and a Model Context Protocol
(MCP) server. The requester wants a minimal footprint that cross-compiles
trivially to darwin, linux, and windows on both amd64 and arm64, and that ships
as one artifact users can drop on a `PATH`.

Two obvious shapes were possible: two separate programs (a linter and a server)
that share a library, or one program that offers both behaviors through
subcommands. Two programs would duplicate build, release, and distribution work,
and would let the two surfaces drift apart.

## Decision

Ship one Go binary, `vale`, built with `CGO_ENABLED=0` so it is a static,
dependency-free executable that cross-compiles with the standard toolchain and
goreleaser. The binary has two faces over one engine:

- `vale lint <paths...>` — the command-line linter.
- `vale mcp` — the stdio MCP server.

Both faces call the same linting engine (`internal/lint` and `internal/linter`).
The MCP server wraps the exact code the CLI uses, so a finding from the CLI and a
finding from the `lint_text` tool are identical. Supporting subcommands (`gen`,
`rules`, `version`) live in the same binary. Command dispatch uses the standard
library `flag` package, not a third-party CLI framework, to keep the footprint
small.

## Consequences

- One build, one release matrix, one artifact to install and document.
- The CLI and the MCP server can never disagree, because they share the engine.
- Pure Go with `CGO_ENABLED=0` gives trivial cross-compilation and a static
  binary with no runtime dependencies.
- The binary carries both surfaces even when a user needs only one; this cost is
  negligible for a linter of this size.
- New behavior is a new subcommand, which keeps the tool coherent.
