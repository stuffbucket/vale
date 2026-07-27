# 2. Hand-rolled JSON-RPC over newline-delimited stdio for MCP

- Status: Accepted

## Context

The `vale mcp` face must speak the Model Context Protocol so an agent can lint
text through the same engine as the CLI. MCP is a JSON-RPC 2.0 protocol. Clients
such as Claude Code launch the server as a subprocess and talk to it over
standard input and output.

An MCP SDK could provide the transport and the type definitions. However, the
server's surface is tiny: it must answer `initialize`, `tools/list`,
`tools/call`, and `ping`, and accept the `notifications/initialized`
notification. Pulling in an SDK for this would add a dependency tree, enlarge the
binary, and couple the project to an external release cadence — all against the
minimal-footprint goal in [ADR 0001](0001-single-pure-go-binary.md).

## Decision

Hand-roll a JSON-RPC 2.0 server over newline-delimited stdio, using only the
standard library (`encoding/json`, `bufio`, `io`). Each request and each response
is one JSON object on its own line. The implementation lives in `internal/mcp`.

- Transport: read one line, decode a `request`, dispatch, encode one `response`
  line. Notifications (no `id`) produce no response.
- Methods: `initialize`, `notifications/initialized`, `ping`, `tools/list`,
  `tools/call`. Unknown methods return JSON-RPC error `-32601`.
- Protocol version: the server reports `2024-11-05`. It echoes the client's
  requested `protocolVersion` when the client sends one.
- Tools: `lint_text` and `list_rules`, each with an input schema.

No MCP SDK is a dependency.

## Consequences

- The MCP server adds no third-party dependency and almost no binary size.
- The surface is small enough to read, test, and fuzz in full; a scripted stdio
  round-trip verifies `initialize` plus a `lint_text` call.
- The project owns the protocol details, so it must track the MCP specification
  by hand if the protocol version changes.
- The transport assumes one JSON object per line. This is simple and robust for
  the stdio subprocess model that MCP clients use.
