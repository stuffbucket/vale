# Lexical Spikes

## The spike

A small, closed set of rare words jumped from a near-zero, decade-stable baseline
to a sudden post-ChatGPT surge. In 15M+ PubMed abstracts, `delve`/`delves`/
`delving` rose ~28x, `underscores` ~13.8x, and `showcasing` ~10.7x
([2406.07016](../findings/references.md)). Across arXiv CS abstracts, `realm`,
`intricate`, `showcasing`, and `pivotal` jumped from a ~2-3% decade baseline
([2404.01268](../findings/references.md)).

Because the human baseline for these words is low, a single occurrence carries
more signal here than for any other lexical category — and most are also vague,
non-plain diction that STE discourages on independent grounds. This is the one
category where per-occurrence flagging is defensible (see
[candidate-rules.md](../remediation/candidate-rules.md), `STE.SlopVocabulary`).

## Why these words, and why the register

The spike is a surface symptom of how the models were trained and tuned. The
[taste-skill](https://github.com/Leonxlnx/taste-skill) research (MIT) documents
the generation-time mechanisms; the ones that shape *diction* (not just length)
are:

- **RLHF and reward shaping.** Post-training alignment rewards confident,
  polished, "professional-sounding" summaries. Praise-heavy, elevated vocabulary
  correlates with the high-quality academic and enterprise text in the reward
  distribution, so the model learns to reach for it. (taste-skill:
  *root-causes/rlhf-and-compute.md*.)
- **Training-data imitation.** The model imitates the register of the polished
  documents it saw most, converging on a narrow "house style" rather than the
  variance of individual human authors. (taste-skill:
  *root-causes/training-data-bias.md*.)

## The catch: markers decay

Marker discriminative power is not stable. After `delve` was publicly called out
in early 2024, its arXiv frequency dropped while unnoticed favored words
(`significant`, `additionally`, `crucial`) kept rising
([2502.09606](../findings/references.md)). Any hardcoded slop wordlist ages and
needs periodic revision — a reason to keep the list small, versioned, and
suppressible rather than authoritative.
