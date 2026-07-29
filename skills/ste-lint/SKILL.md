---
name: ste-lint
description: >-
  Check, simplify, or rewrite text into Simplified Technical English (ASD-STE100)
  with the vale linter. Use when the user asks to lint, review, tighten, or
  rewrite prose, docs, READMEs, procedures, or agent instructions for STE — short
  sentences, active voice, no contractions, plain vocabulary — or to catch "AI
  slop," fix a document with a model, or run STE checks over a repository. Wraps
  the vale CLI, its MCP server, and its model-eval mode.
allowed-tools: Bash(vale *), Bash(${CLAUDE_SKILL_DIR}/scripts/*)
---

# vale — Simplified Technical English

`vale` checks and improves text against Simplified Technical English (STE) rules.
It is an approximation of ASD-STE100 (not certified) — treat its output as
guidance, not a certificate.

## When to use

- Lint or review text, docs, READMEs, or procedures for STE.
- Rewrite prose into STE, or make instructions shorter and more direct.
- Catch "AI slop" (LLM-ish diction, hedging, restatement) — add `--slop`.
- Fix a document automatically with a model — `vale --fix`.

## Core commands

Linting is the default action. The default output groups findings by rule and is
compact; add `--format json` when you need to parse it.

```sh
vale README.md                 # concise findings
vale docs/ --format json       # machine-readable
vale --slop notes.md           # include the opt-in slop rules
```

If `vale` is not on `PATH`, use `${CLAUDE_SKILL_DIR}/scripts/lint.sh <path>` (it
falls back to `go run`). Install it with `brew install stuffbucket/tap/vale`.

Exit codes: `0` clean at the gate, `1` findings at or above the gate, `2` a tool
error. Pass `--audit` to always exit `0` (collect findings without failing).

## Rewrite, do not only report

When asked to rewrite, fix the findings in place and keep the meaning exact:
expand contractions, shorten sentences (20 words or fewer in steps), use the
active voice, put one instruction in one sentence, drop `-ing` forms and phrasal
verbs, and use the approved word from each finding's `hint`. Re-run until clean.
For an automatic rewrite, use `vale --fix <file>` — it prints the corrected
document to stdout, or to `--output <file>`.

## Going deeper (load only when you need it)

- Rules, severities, and the rewrite checklist — [references/rules.md](references/rules.md)
- Config layers, vocabulary allow/deny, slop, inline suppression — [references/config.md](references/config.md)
- MCP server (session vocabulary) and model eval — [references/mcp-and-eval.md](references/mcp-and-eval.md)
- Starter config to copy — [templates/vale-ste.yml](templates/vale-ste.yml)
