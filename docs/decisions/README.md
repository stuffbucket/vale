# Architecture decision records

This directory holds the architecture decision records (ADRs) for vale. Each ADR
captures one decision, its context, and its consequences. The format is short:
Title, Status, Context, Decision, Consequences.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-single-pure-go-binary.md) | A single, pure-Go binary with two faces | Accepted |
| [0002](0002-mcp-transport-handrolled-jsonrpc.md) | Hand-rolled JSON-RPC over newline-delimited stdio for MCP | Accepted |
| [0003](0003-config-format-yaml.md) | YAML configuration via gopkg.in/yaml.v3 | Accepted |
| [0004](0004-self-referential-vocab-generation.md) | Self-referential vocabulary generation from the OpenSTE wordset | Accepted |
| [0005](0005-rule-model.md) | Rule model: one rule per file, stable IDs, three severities | Accepted |
| [0006](0006-ast-markdown-parsing-and-lsp.md) | AST-based Markdown parsing (goldmark), path scoping, and an LSP surface | Accepted (stages 1-2) |
| [0007](0007-inline-suppression-directives.md) | Inline rule suppression via HTML-comment directives | Accepted |
| [0008](0008-llm-fix-and-model-config.md) | LLM-assisted fix, shared model config, and progress UI | Accepted |

## Adding an ADR

1. Copy the format of an existing record.
2. Use the next number in sequence.
3. Set the status (`Proposed`, `Accepted`, or `Superseded`).
4. Add a row to the table above.
