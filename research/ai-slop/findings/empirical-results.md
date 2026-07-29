# Empirical Results

Concrete numbers from the sweep. Full citations in
[references.md](references.md). Papers are grouped by what they measured.

## Prevalence and frequency shifts (the core evidence)

- **[2406.07016]** — Post-2023 frequency spikes across 15M+ PubMed abstracts:
  `delves` ratio ~28.0, `underscores` ~13.8, `showcasing` ~10.7. Estimate:
  **≥13.5%** of 2024 biomedical abstracts were LLM-processed. Excess vocabulary is
  ~66% verbs, ~14% adjectives — the signal is statistical, not per-text.
- **[2404.01268]** — Word-frequency maximum-likelihood framework: **~17.5%** of CS
  abstracts LLM-modified by Feb 2024. `realm`, `intricate`, `showcasing`,
  `pivotal` jumped from a decade-long **~2-3%** baseline — which doubles as the
  irreducible false-positive floor.
- **[2403.07183]** — Corpus mixture model: **6.5-16.9%** of post-ChatGPT peer
  reviews LLM-modified. Signal carried by a few adjectives: `commendable` ~9.8x,
  `meticulous` ~34.7x, `intricate` ~11.2x. Explicitly undetectable at the
  individual level.
- **[2403.16887]** — ~24 praise words spiked in 2023 (`intricate` +117%,
  `meticulously` +137%) vs ≤5-11% for neutral controls. Co-occurrence of two
  strong markers rose **~7-fold** — multi-marker density is far more specific than
  single words.
- **[2502.09606]** — After `delve` was publicly called out, its arXiv frequency
  **fell** while `significant`/`additionally` kept rising. Markers decay; corpus
  tracking stays viable, text-by-text detection is "perhaps impossible."
- **[2505.12218]** — Post-ChatGPT arXiv abstracts show **+12-13%** LLM-adjective
  density and up to **+8%** adverb density, plus higher sentiment/subjectivity.

## Hedging, verbosity, sycophancy

- **[2509.24202]** — Supplies a hedge lexicon and impersonal-it modal patterns,
  but frames hedging as legitimate human-aligned uncertainty. Only clustering is a
  signal.
- **[2411.07858]** — "Verbosity compensation" occurs at a **13-74%** rate across
  five subtypes; the `>3-token` detector is QA-tuned and does **not** transfer to
  prose.
- **[2404.04475]** — Verbosity bias is strong enough that a model can nearly
  triple its win rate by padding; length is a confound, not a quality signal.
- **[2308.03958]** — Sycophancy grows with scale; behavioral, no surface-text
  markers a single-document linter can read.

## Lexical diversity and detection methods (mostly non-transferable)

- **[2308.07462]** — AI corpora had ~25-50% fewer distinct lemmas and lower RTTR
  (e.g. Computing 29.75 vs 45.43). Needs thousands of tokens; authors call it
  "very preliminary" (GPT-3.5 only).
- **[2301.11305] DetectGPT** — Probability-curvature detector (AUROC up to 0.95).
  Needs token-level log-probs from the model — not a surface feature a linter sees.
- **[2304.14276]** — ChatGPT essays run **higher on nominalization**, **lower on
  hedges**, and use **fewer** discourse connectives than humans (refuting the
  "AI overuses however/moreover" myth).
- **[2401.16587]** — LIWC study: ChatGPT skews to `you`/`we`, suppresses `I`, and
  overuses politeness/prosocial tokens — dialogue-specific, register-bound.
- **[2302.00937]** — Sentence-splitting readability: prefer 2-way over 3-way
  splits. A simplification-quality metric, not a slop tell.
- **[1605.02457] Thing Explainer CNL** — Restricts text to the ~1,000 most common
  words with 13 morphology rules. A vocabulary-commonness metric whose lexical
  signal points **opposite** to precise technical writing.

## Model-family attribution

- **[2503.01659]** — Attributes text to **Claude, OpenAI, Gemini, Llama** at
  **0.9988** precision, **0.0004** false-positive rate, via a three-classifier
  ensemble that commits only when unanimous. Models keep "distinct and consistent
  stylistic fingerprints, even when prompted to write in different writing
  styles." Confirms per-family detection is real — but via learned classifiers,
  not a surface wordlist, so a linter's word-match is a weaker proxy.

## Repetition and reinvention

- **[1904.09751]** — Likelihood-maximizing decoding yields "bland and strangely
  repetitive" text; the diagnosis behind nucleus sampling.
- **[2206.02369]** — The **self-reinforcement effect**: after a sentence repeats,
  the probability of repeating it again rises — repetition compounds. This is the
  mechanism behind "the same concept over and over."
- **[2012.14660]** — Formalizes it as the high-inflow problem with an Average
  Repetition Probability metric.
- **[2410.13497]** — Identifies specific "repetition neurons" that drive the
  behavior.
- Practical linter metrics from this literature: **rep-n** (repeated-n-gram
  fraction), **distinct-n** (unique-n-gram ratio), **self-BLEU** (corpus-level
  self-similarity), and sentence-level **shingling + Jaccard** for near-duplicate
  detection.
