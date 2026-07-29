# 0007. Inline rule suppression via HTML-comment directives

Status: Accepted

## Context

Clean reports need a way to silence a rule on a specific line or region — for
example, the research docs quote slop words on purpose, and generated tables hold
terms STE would flag. The requirement is line-level at minimum, and the notation
must reuse an established convention rather than invent one.

## Decision

Reuse HTML-comment directives, which every Markdown-oriented linter already uses
and which vale's goldmark AST already treats as non-prose (so the directives
never themselves produce findings). The keyword is `vale`, and the grammar unions
two established families:

**Region toggles (Vale's own syntax):**

- `<!-- vale off -->` … `<!-- vale on -->` — all rules off, then on.
- `<!-- vale STE.PassiveVoice = NO -->` … `<!-- vale STE.PassiveVoice = YES -->`
  — one rule off, then on.

**Line-scoped (markdownlint's established verbs):**

- `<!-- vale disable-line [STE.Rule …] -->` — ignore the same line.
- `<!-- vale disable-next-line [STE.Rule …] -->` — ignore the next line.

With no rule ids, a line directive suppresses every rule; with one or more
space-separated ids, only those. Nothing here is novel: the prefix and
`off`/`on`/`= NO`/`= YES` come from Vale, the `disable-line` / `disable-next-line`
verbs from markdownlint.

## Implementation

- `internal/lint/suppress.go` scans the raw source line by line for
  `<!-- vale … -->` comments, tracks the running region state (all-off and
  per-rule-off), and records, per line, which rule ids (or `*`) are suppressed.
- The linter (`internal/linter`) drops any finding whose `(line, ruleID)` is
  suppressed after the engine runs. Region state is line-granular, which is
  sufficient for the stated need.

## Consequences

- Line- and region-level suppression with a familiar notation; no config churn.
- Directives are inert prose-wise (already masked), so they cannot self-trigger.
- Line granularity means an inline `off` suppresses its whole line; acceptable per
  the requirement. A finer column-granular pass can come later if needed.
