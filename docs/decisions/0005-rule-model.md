# 5. Rule model: one rule per file, stable IDs, three severities

- Status: Accepted

## Context

Vale enforces a set of Simplified Technical English rules. The rules are diverse:
some inspect single tokens (contractions), some inspect whole sentences
(sentence length, one instruction), and some inspect verb phrases (passive
voice, phrasal verbs). The rule set must be easy to extend, easy to test, and
stable enough that users can refer to a rule by ID in configuration and in
suppression.

The rules also run over Markdown as well as plain text. A naive tokenizer would
flag code spans, fenced code, and link destinations, and would miscount headings
and list items as prose.

## Decision

Model each rule as a small type that satisfies the `lint.Rule` interface, with
one rule per file under `internal/rules` and a table-driven `_test.go` beside it.

- **Stable IDs** in the `STE.*` namespace: `STE.SentenceLength`,
  `STE.Contractions`, `STE.PassiveVoice`, `STE.IngForms`, `STE.PhrasalVerbs`,
  `STE.OneInstruction`, `STE.Vocabulary`. IDs are the contract for config and
  reporting.
- **Three severities**: `error`, `warning`, `suggestion`. Each rule declares a
  default severity; config can override it or disable the rule.
- **Table-driven tests** per rule, with good cases and bad cases, so behavior is
  pinned and regressions are visible.
- A **Markdown-aware tokenizer** (`internal/lint`) masks fenced code, inline
  code, and link destinations, and marks headings and list items as distinct
  block kinds so rules can change behavior for them.
- **Sentence length** uses a split limit: list items (procedures) get the
  smaller limit (20 words by default) and description sentences get the larger
  limit (25 words by default). Headings are exempt. Both limits are configurable.

## Consequences

- A new rule is a new file plus a table-driven test plus one line of
  registration; the pattern is uniform and reviewable.
- Stable IDs let users disable or re-severity a rule and let editors group
  findings.
- The Markdown-aware tokenizer keeps findings on prose, not on code or links,
  which avoids a class of false positives.
- The split sentence-length limit matches STE practice, where procedure steps are
  held to a tighter bound than descriptive text.
- The three-severity model gives CI a clear gate: fail on `error` by default,
  raise the gate to include warnings when a project wants a stricter bar.
