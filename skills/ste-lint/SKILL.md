---
name: ste-lint
description: >-
  Lint or rewrite text into Simplified Technical English (ASD-STE100) and review
  documentation for STE compliance. Use this skill when the user asks to check,
  simplify, or rewrite prose so that it follows Simplified Technical English —
  short sentences, no contractions, active voice, and approved vocabulary — or
  to review docs, READMEs, procedures, or instructions for STE issues. Wraps the
  `vale` linter and its MCP server.
---

# STE lint

This skill uses the `vale` binary to check and improve text against Simplified
Technical English (STE) rules. Vale is an approximation of ASD-STE100 (not
certified), so treat its output as guidance, not a certificate.

## When to use

- The user asks to lint, check, or review text or docs for STE compliance.
- The user asks to rewrite prose into Simplified Technical English, or to make
  instructions shorter, clearer, and more direct.
- You are writing procedures, READMEs, or agent instructions and want them to
  pass an STE gate.

## How to run the linter

Prefer the JSON format; it is stable and easy to parse.

```sh
vale lint <path> --format json
```

If `vale` is not on `PATH`, use the bundled wrapper, which falls back to
`go run`:

```sh
scripts/lint.sh <path>
```

You can also lint several files or a directory:

```sh
vale lint docs/ README.md --format json
```

Useful flags:

- `--min-severity error|warning|suggestion` — the exit-code gate. Exit `0` means
  clean at the gate, `1` means findings at or above the gate, `2` means a usage
  or runtime error.
- `--markdown auto|on|off` — force Markdown-aware tokenizing. `auto` decides from
  the file ending.
- `--strict-vocabulary` — also flag unapproved words that have no direct
  replacement. Off by default to avoid a flood of low-value findings.

### Interpreting findings

Each JSON finding has:

- `ruleId` — one of `STE.SentenceLength`, `STE.Contractions`, `STE.PassiveVoice`,
  `STE.IngForms`, `STE.PhrasalVerbs`, `STE.OneInstruction`, `STE.Vocabulary`.
- `severity` — `error`, `warning`, or `suggestion`.
- `message` — what is wrong.
- `hint` — how to fix it (often the approved replacement word).
- `line`, `col`, `endLine`, `endCol`, `match` — the exact span.

Fix errors first, then warnings, then suggestions. Re-run the linter after each
pass until it is clean at the gate you care about.

## How to use the MCP server

For interactive or repeated linting without touching the filesystem, connect the
MCP server:

```sh
vale mcp
```

It speaks JSON-RPC 2.0 over stdio, one message per line, protocol version
`2024-11-05`. Register it in an MCP client:

```json
{
  "mcpServers": {
    "vale": {
      "command": "vale",
      "args": ["mcp"]
    }
  }
}
```

Call the `lint_text` tool with a `text` argument (required). Optional arguments:
`filename` (a `.md` name turns on Markdown mode), `markdown` (boolean, overrides
the file name), and `minSeverity`. Call `list_rules` to see every rule with its
default severity. The `lint_text` result is a text block: a one-line summary,
then a JSON block with the `findings` array and a `summary` count by severity.

## Rewrite workflow

When you rewrite flagged text into Simplified Technical English:

1. **Expand contractions** (`STE.Contractions`). Write `do not` for `don't`,
   `it is` for `it's`, and `cannot` for `can't`.
2. **Shorten sentences** (`STE.SentenceLength`). Keep list items and procedure
   steps to 20 words or fewer, and description sentences to 25 words or fewer.
   Split one long sentence into two short ones.
3. **Use the active voice** (`STE.PassiveVoice`). State who does the action:
   turn "the valve was closed" into "close the valve" or "the operator closed
   the valve".
4. **Write one instruction per sentence** (`STE.OneInstruction`). Give each step
   a single action, in the imperative for procedures.
5. **Avoid `-ing` verb forms in instructions** (`STE.IngForms`). Use the plain
   imperative.
6. **Replace phrasal verbs** (`STE.PhrasalVerbs`) with one clear verb.
7. **Use approved vocabulary** (`STE.Vocabulary`). Replace each flagged word with
   the approved word from the finding's `hint`.

After rewriting, run `vale lint <path> --format json` again and confirm the
findings are gone. Do not change the meaning of the text; keep the technical
content exact.
