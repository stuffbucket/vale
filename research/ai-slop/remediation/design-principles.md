# Design Principles

The constraints every `STE.Slop*` rule shares, and how detection-time linting
fits alongside generation-time remediation.

## 1. Corpus-level evidence, per-line tool

The evidence is population-level; vale fires per line. Bridge the gap by framing:
a slop rule says "this word is vague; prefer a concrete term," not "this line is
AI." Prefer **density thresholds over a scope** and **multi-marker co-occurrence**
over any single generic word match — co-occurrence of two praise markers is ~7x
more specific than one ([2403.16887](../findings/references.md)).

## 2. Default to warning, always suppressible

Every slop rule defaults to `warning` or `suggestion` severity (never `error`),
is individually toggleable, and honors the existing suppression and
technical-terms allowlist. The lexical watchlist ages, so keep it small, versioned,
and easy to override — do not treat it as authoritative.

## 3. Reuse the existing machinery

Slop rules should compose with what vale already has, not duplicate it:

- The **technical-terms allowlist** exempts domain vocabulary from
  nominalization and vocabulary checks.
- `STE.ImpersonalItModalHedge` **complements** the passive-voice rule rather than
  overlapping it.
- `STE.VaguePraiseSubstitution` reuses the substitution machinery already built
  for `STE.Vocabulary`.

## 4. Detection is only half the problem

vale detects slop that exists. It cannot stop a model from producing it. The
generation-time half is covered by
[taste-skill](https://github.com/Leonxlnx/taste-skill) (MIT), whose remediation
research is worth pairing with vale in a writing pipeline:

- **Prompt engineering** — structural binding, XML-tagged prompts, and
  verification/self-grading loops that reduce padding and shortcutting at the
  source. (taste-skill: *remediation/prompt-engineering.md*.)
- **Parameter tuning** — temperature/top-p and thinking-level settings that affect
  output length and depth. (taste-skill: *remediation/parameter-tuning.md*.)

The clean division of labor: **generate with taste-skill's techniques, then lint
the result with vale.** Neither alone is sufficient — a well-prompted model still
drifts into the register documented in
[root-causes/formulaic-register.md](../root-causes/formulaic-register.md), and a
linter cannot add substance a lazy generation left out.

## 5. Ship in tiers

Build in the order of [candidate-rules.md](candidate-rules.md): the two
deterministic Tier-1 rules (`SlopVocabulary`, `ImpersonalItModalHedge`) and the
`NominalizationDensity` metric first, since they carry the best evidence-to-false-
positive ratio and align with STE's existing goals. Treat the density and Tier-3
rules as opt-in experiments gated behind config until validated on real corpora.
