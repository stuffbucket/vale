# AI Slop: markers, evidence, and rules

A structured analysis of what makes LLM-generated prose read as "slop," which of
those markers a Simplified Technical English linter can act on, and — just as
important — which popular intuitions the evidence refutes.

The findings come from a 16-paper arXiv sweep (2016-2025) on LLM text detection,
lexical-frequency shifts in academic writing, hedging, verbosity, and controlled
natural language. See [findings/references.md](findings/references.md) for the
full list.

## The one result that governs everything

Every discriminative finding in the literature is **corpus-level, not per-text**.
Multiple papers explicitly disclaim per-document inference; one
([2502.09606](findings/references.md)) calls text-by-text detection "perhaps
impossible." A linter fires per line, so vale must frame every slop rule as a
**plain-language / style nudge**, never as a verdict that a passage was written by
AI. This single constraint decides which markers become rules and which stay as
advisories.

## Directory structure

### Root Causes
Why slop looks the way it does — the generation mechanisms upstream of the text.

- [Lexical Spikes](root-causes/lexical-spikes.md) — how a small set of rare words
  jumped from near-zero to a post-2023 spike, and the economic/training pressures
  behind the register.
- [Formulaic Register](root-causes/formulaic-register.md) — praise adjectives,
  hedging, and verbosity padding as a learned house style.
- [Homogenization and Inversion](root-causes/homogenization-and-inversion.md) —
  low-variance, nominalization-heavy output, and the crucial fact that several of
  these signatures **invert** for well-formed STE.

### Findings
- [Empirical Results](findings/empirical-results.md) — the concrete numbers per
  paper (frequency ratios, prevalence estimates, effect sizes).
- [References](findings/references.md) — the 16 arXiv papers plus external
  resources (taste-skill, OpenSTE, ASD-STE100).

### Remediation
- [Candidate Rules](remediation/candidate-rules.md) — nine candidate `STE.Slop*`
  rules mapped to vale's rule types, each with its evidence and false-positive
  risk.
- [Anti-Rules](remediation/anti-rules.md) — rules the evidence tells us **not** to
  build, including the connective-overuse myth and the STE inversions.
- [Design Principles](remediation/design-principles.md) — corpus-vs-per-line,
  default severity, allowlist and suppression, and the generation-time complement.
