# Candidate Rules

Nine candidate rules, mapped to vale's four rule types (existence, substitution,
occurrence, metric) and ranked by build priority — safest and highest-value
first. Each is grounded only in the cited papers. See
[design-principles.md](design-principles.md) for the shared constraints (warning
severity, allowlist exemption, suppression) and
[anti-rules.md](anti-rules.md) for what **not** to build.

Default all of these to **warning** or **suggestion**, individually toggleable,
under an opt-in `STE.Slop*` family. None should autofix.

## Tier 1 — deterministic, low false-positive

### `STE.SlopVocabulary` (existence)
A closed watchlist of low-baseline "spike" words that surged post-2023 and are
also vague, non-plain diction STE discourages anyway: `delve`/`delves`/`delving`,
`underscore(s)`, `showcasing`, `intricate`, `meticulous(ly)`, `pivotal`, `realm`,
`elucidate`, `groundbreaking`, `seamlessly`, `commendable`, `unparalleled`,
`multifaceted`, `leverage`, `harness`. The one lexical category where per-hit
flagging is defensible (human baseline ~2-3%). Frame as a style nudge, never as
proof of AI authorship. Evidence: 2406.07016, 2502.09606, 2404.01268, 2403.16887,
2403.07183, 2505.12218. FP risk: **low-moderate** (legitimate `intricate
mechanism`, `pivotal joint`; markers decay — keep the list versioned).

### `STE.ImpersonalItModalHedge` (existence)
Regex for the impersonal-it + modal + reporting-verb padding pattern:
`it (can|could|may|might) be (argued|said|assumed|shown) that`, `it is possible
that`, `there is a possibility that`. Wordy, evasive, passive-adjacent — its
rewrite (a direct declarative) is exactly what STE wants. Evidence: 2509.24202,
2411.07858. FP risk: **low** (specific, uncommon in well-formed instructions).

### `STE.VaguePraiseSubstitution` (substitution)
Where a flagged non-plain word has a meaning-preserving plain replacement, offer
it — the STE-native way to act on the slop lexicon: `utilize` → `use`, `leverage`
→ `use`. Evidence: 1605.02457, 2406.07016, 2505.12218. FP risk: **moderate** —
keep to the low-risk subset; evaluative swaps (`comprehensive` → `complete`) shift
nuance, so suggest-only, never auto-applied.

## Tier 2 — density metrics, not per-hit flags

### `STE.NominalizationDensity` (occurrence)
Per-sentence ratio of derived-noun suffixes (`-tion`, `-sion`, `-ment`, `-ance`,
`-ence`, `-ness`, `-ity`). LLM text runs high **and** STE penalizes it, so the
signal and the tool agree (the "frozen verbs" case). Exempt the technical-terms
allowlist; nudge toward the direct verb; no autofix. Evidence: 2304.14276. FP
risk: **moderate-high** (many legitimate technical nouns carry these suffixes —
tune the threshold).

### `STE.EvaluativeAdjectiveDensity` (occurrence)
Density of praise adjectives and `-ly` evaluative adverbs (`commendable`,
`meticulous`, `notable`, `noteworthy`, `innovative`, `versatile`, `comprehensive`,
`invaluable`, `robust`, `remarkable`, `exceptional`; `thoroughly`, `notably`,
`effectively`). The evidence is explicit that concentration, not any single hit,
is the signal, and co-occurrence of 2+ is far more specific. Advisory count, no
autofix. Evidence: 2403.07183, 2403.16887, 2505.12218, 2406.07016, 2502.09606. FP
risk: **high per-occurrence, low-moderate as density**.

### `STE.HedgeDensity` (occurrence)
Clustering of hedge tokens (`may`, `might`, `could`, `likely`, `possibly`,
`perhaps`, `seems`, `appears`, `typically`, `generally`) around a single
assertion. Individual hedges are legitimate; only abnormal packing is a signal.
Soft warning, human review. Evidence: 2509.24202, 2304.14276. FP risk: **high for
a flat blocklist** — flag only multi-hedge clusters.

## Tier 3 — narrow or weak; build last or not at all

### `STE.VerbosityCompensationPatterns` (existence)
Only the narrowest structural patterns: a vague quantifier where a value is
expected, a comma-separated candidate list offered as one answer, question-echo.
Evidence: 2411.07858, 2404.04475. FP risk: **high, hard to operationalize** — the
source detector is QA-tuned and does not transfer to prose; length must never be a
standalone flag.

### `STE.CommonWordOveruse` (metric, off by default)
Whole-document density of ordinary words that carry most of the statistical signal
but are normal technical vocabulary (`comprehensive`, `crucial`, `significant`,
`particularly`, `notably`, `additionally`, `within`, `across`, `effectively`,
`enhance`, `capabilities`, `potential`, `findings`). Corpus-level advisory only.
Evidence: 2406.07016, 2502.09606, 2404.01268. FP risk: **very high** — the
baseline (~2-3%) is the false-positive floor; never a per-line flag.
