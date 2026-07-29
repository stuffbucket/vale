# 0009. Session-adaptive vocabulary and slop-triggering hook

Status: Accepted

## Context

Running as an MCP server, vale should adapt to the session: an agent that
approves a project term should stop seeing it flagged for the rest of the
session, and the learning should be shareable with the CLI. Separately, vale is
more effective when it triggers on slop itself rather than waiting for the agent
to remember to call a lint tool — but the MCP surface should stay tight.

## Decision

1. **A shared vocab store.** `.vale-ste.vocab.yml` (overridable) is a dedicated,
   vale-managed config fragment (`vocabulary.allow` / `deny`). It is a discovered
   config layer that sits just below an explicit `--config`, so the MCP, the CLI,
   and the hook all honor learned terms. `config.UpdateVocabStore` rewrites it
   wholesale (vale owns it); `config.ReadVocabStore` reads it.
2. **Session-mode MCP.** `mcp.NewSessionServer` loads config from the working
   directory and rebuilds its linter after each vocab change. The
   `update_vocabulary` tool (allow / deny arrays) persists terms and reloads, so
   later `lint_text` / `fix_text` calls in the same session honor them. Scope is
   set by the store path: the default `.vale-ste.vocab.yml` shares learning
   across a project; `--vocab-store /tmp/vale-<session>.yml` scopes it to one
   session.
3. **The slop watchlist honors the allow set.** `STE.SlopVocabulary` now takes
   the allowed set, so approving a watchlist word actually silences it.
4. **A lint-on-write hook, not more MCP tools.** The plugin's `PostToolUse` hook
   (`hooks/lint-on-write.py`) lints text files the agent writes with
   `vale --slop --audit` and returns findings as `additionalContext`. This makes
   vale trigger on slop automatically. The `.mcp.json` stays one tight server; a
   new `--slop` CLI flag lets the hook enable the family.

## Consequences

- The MCP adapts within a session and can persist learning across sessions,
  keyed by the store path.
- Learned terms are honored uniformly by the MCP, the CLI, and the hook.
- vale catches slop proactively without bloating the MCP tool list.
