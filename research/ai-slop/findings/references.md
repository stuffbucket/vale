# References

The 16 arXiv papers from the sweep, plus the external resources vale's slop
research builds on. Papers are ordered by relevance to a linter.

## Directly actionable

- **[2406.07016]** Kobak et al. — *Delving into LLM-assisted writing in biomedical
  publications through excess vocabulary.* The anchor paper: post-2023 style-word
  spikes, ≥13.5% prevalence estimate. https://arxiv.org/abs/2406.07016
- **[2403.16887]** — *ChatGPT "contamination": estimating the prevalence of LLMs
  in the scholarly literature.* The praise-word watchlist and multi-marker
  co-occurrence result. https://arxiv.org/abs/2403.16887
- **[2403.07183]** Liang et al. — *Monitoring AI-Modified Content at Scale.*
  Adjective-driven corpus mixture model on peer reviews.
  https://arxiv.org/abs/2403.07183
- **[2304.14276]** — *AI, write an essay for me: human vs ChatGPT.* Nominalization
  density up, hedges down, connectives down. https://arxiv.org/abs/2304.14276
- **[2509.24202]** — *Can Large Language Models Express Uncertainty Like Human?*
  Hedge lexicon and impersonal-it modal patterns. https://arxiv.org/abs/2509.24202

## Corpus-level context (metric-only or cautionary)

- **[2404.01268]** — *Mapping the Increasing Use of LLMs in Scientific Papers.*
  https://arxiv.org/abs/2404.01268
- **[2502.09606]** — *Human-LLM Coevolution: Evidence from Academic Writing*
  (marker decay). https://arxiv.org/abs/2502.09606
- **[2505.12218]** — *Examining Linguistic Shifts in Academic Writing Before and
  After the Launch of ChatGPT.* https://arxiv.org/abs/2505.12218
- **[2411.07858]** — *Verbosity != Veracity.* https://arxiv.org/abs/2411.07858
- **[2404.04475]** — *Length-Controlled AlpacaEval* (verbosity bias).
  https://arxiv.org/abs/2404.04475
- **[2308.07462]** — *Playing with words: vocabulary and lexical diversity.*
  https://arxiv.org/abs/2308.07462
- **[2401.16587]** — *A Linguistic Comparison between Human and ChatGPT-Generated
  Conversations.* https://arxiv.org/abs/2401.16587
- **[2308.03958]** — *Simple synthetic data reduces sycophancy.*
  https://arxiv.org/abs/2308.03958

## Detection methods and controlled language (background)

- **[2301.11305]** — *DetectGPT.* https://arxiv.org/abs/2301.11305
- **[2302.00937]** — *The Fewer Splits are Better* (sentence-split readability).
  https://arxiv.org/abs/2302.00937
- **[1605.02457]** — *The Controlled Natural Language of Randall Munroe's Thing
  Explainer.* https://arxiv.org/abs/1605.02457

## External resources

- **[Leonxlnx/taste-skill](https://github.com/Leonxlnx/taste-skill)** (MIT) —
  "gives your AI good taste; stops boring, generic slop." Its `research/laziness/`
  covers the *generation-time* half of slop (RLHF/compute economics, training-data
  placeholder bias, cognitive shortcuts, output limits; remediation via prompt
  engineering and parameter tuning). vale's research reuses its folder structure
  and cites its root-cause analysis where diction is concerned. Its references:
  EmotionPrompt (Microsoft Research), LazyBench, and the 2025 controlled-laziness
  experiments.
- **[OpenSTE](https://github.com/openste/openste)** (MIT) — the open Simplified
  Technical English wordset vale generates its `STE.Vocabulary` rule from.
- **[ASD-STE100](https://www.asd-ste100.org/)** — the Simplified Technical English
  standard. Its Technical Names / Technical Verbs mechanism is why a domain
  allowlist is STE-compliant, not a workaround.
