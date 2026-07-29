# MCP server and model eval

## MCP server

`vale mcp` starts a stdio MCP server (JSON-RPC 2.0, one message per line,
protocol `2024-11-05`). Register it in an MCP client:

```json
{ "mcpServers": { "vale": { "command": "vale", "args": ["mcp"] } } }
```

Tools:

- **`lint_text`** — check a string. Args: `text` (required), `filename`
  (`.md` turns on Markdown), `markdown` (boolean), `minSeverity`. The result is a
  one-line summary plus the concise findings grouped by rule.
- **`list_rules`** — every rule with its default severity.
- **`fix_text`** — rewrite text with a model so it resolves its findings. Args:
  `text` (required), `filename`, `model`, `temperature`, `maxTokens`.
- **`update_vocabulary`** — learn project vocabulary for the session. Args:
  `allow` (approve terms) and `deny` (re-check terms). It persists to the learned
  store (`$XDG_STATE_HOME/vale-ste/vocab.yml`) and rebuilds the linter, so later
  `lint_text` and `fix_text` calls in the session honor the terms. When a
  legitimate project word keeps being flagged, approve it with this tool rather
  than suppressing each occurrence.

The server runs in session mode: it adapts to the session and persists learning.
Scope the store with `vale mcp --vocab-store <path>` (a per-session path for
session-only learning; the default is shared across sessions).

## Fixing a document

`vale --fix <file>` lints the file, sends the text and findings to a model on the
configured endpoint (default `http://localhost:4141`), and prints the corrected
document to stdout — or to `--output <file>`. Flags: `--model`, `--temperature`,
`--max-tokens`. Configure defaults in the `model` config block.

## Measuring slop across models

`vale eval` drives an OpenAI-compatible endpoint, prompts each model, lints every
reply with the slop rules, and reports slop density per model and family
(families come from `/v1/models` `owned_by`).

```sh
vale eval                                   # discover and test every model
vale eval --models "claude-sonnet-5,gpt-5.5,gemini-2.5-pro" --temperature 0.2
```

Read the `slop` (STE.Slop findings per 100 words) column as a directional
comparison, not a precise leaderboard — the markers are low-baseline and noisy at
small sample sizes.
