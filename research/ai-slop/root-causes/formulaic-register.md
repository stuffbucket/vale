# Formulaic Register

Beyond the rare "spike" words, slop has a formulaic register: a predictable
distribution of praise, hedging, and padding. Unlike the spike words, these
signals are **distributional** — reliable only in aggregate, dangerous per
occurrence.

## Positive-evaluative praise

The strongest corpus-level tell is a cluster of praise adjectives and `-ly`
adverbs. In AI-conference peer reviews, `commendable` rose ~9.8x, `meticulous`
~34.7x, and `intricate` ~11.2x ([2403.07183](../findings/references.md)). A
separate study tracked ~24 praise words spiking in 2023 (`intricate` +117%,
`meticulously` +137%) against only 5-11% movement in neutral control words, and
found that the **co-occurrence of two markers** (`intricate` + `meticulous`) rose
~7-fold — multi-marker density is far more specific than any single word
([2403.16887](../findings/references.md)).

Every one of these is a legitimate ordinary word. The signal is the elevated
*concentration* of them, which is why the matching rule must be a density metric,
never a per-hit flag ([candidate-rules.md](../remediation/candidate-rules.md),
`STE.EvaluativeAdjectiveDensity`).

## Hedging and low commitment

LLMs lean on a hedge lexicon (`may`, `might`, `could`, `likely`, `possibly`,
`seems`, `appears`, `typically`) and impersonal patterns
(`it can be argued that`). But the hedging paper frames these as **legitimate,
human-aligned uncertainty signals** ([2509.24202](../findings/references.md)): a
flat hedge blocklist would fire on correct conditional statements in safety and
spec writing. Only abnormal *clustering* of hedges around one assertion is a
usable soft signal. The impersonal-it modal pattern, by contrast, is a specific
wordy construction and is safe to flag directly.

## Verbosity and padding

LLMs pad answers that could be compressed (a 13-74% "verbosity compensation" rate)
through five subtypes: hedging, echoing the prompt, enumerating multiple candidate
answers, unrequested detail, and format wrapping
([2411.07858](../findings/references.md)). But length is a **confound, not a
defect** ([2404.04475](../findings/references.md)): automatic evaluators prefer
longer outputs regardless of content, and legitimate reference docs and STE lists
are long by design. Only the narrowest structural patterns (a vague quantifier
where a value is expected; a comma-separated candidate list offered as one answer)
are candidate deterministic rules. Length itself must never be a standalone flag.
