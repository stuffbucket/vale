# Model Signatures: OpenAI and Anthropic

Coverage for the well-known lexical spikes that flag specific model families. The
goal for a linter is **family-flavored vocabulary hygiene**, not authorship
attribution — see the caveat at the end.

## Attribution is feasible (but not from a word match)

*Detecting Stylistic Fingerprints of Large Language Models*
([2503.01659](../findings/references.md)) shows that text can be attributed to
four families — **Claude, OpenAI, Gemini, Llama** — at 0.9988 precision with a
0.0004 false-positive rate, using an ensemble of three trained classifiers that
only commit when unanimous. Models "retain distinct and consistent stylistic
fingerprints, even when prompted to write in different writing styles." So the
signal is real and per-family separable — but the paper achieves it with learned
classifiers over full style, not a surface wordlist. A linter's word-match is a
much weaker proxy; treat it accordingly.

## OpenAI / ChatGPT family (peer-reviewed lexical evidence)

The excess-vocabulary work ([2406.07016](../findings/references.md)) is
ChatGPT-era and is the strongest lexical evidence. Its authors publish the full
**~900 excess words** (2013-2024) as a CSV in
[berenslab/llm-excess-vocab](https://github.com/berenslab/llm-excess-vocab) — the
authoritative source to draw the OpenAI-family list from. Highest-signal:
`delve`/`delves`/`delving`, `underscore(s)`, `showcasing`, `intricate`,
`meticulous(ly)`, `pivotal`, `realm`, `boasts`, `garner`, `interplay`, `tapestry`,
`testament`.

[Wikipedia's *Signs of AI writing*](https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing)
tracks the drift by era (descriptive, not peer-reviewed, but well-curated):

- **GPT-4 (2023-mid-2024):** delve, boasts, tapestry, meticulous, garner,
  interplay, nestled, testament.
- **GPT-4o (mid-2024-mid-2025):** align with, enhance, fostering, showcasing.
- **GPT-5 (mid-2025+):** emphasizing, enhance, highlighting, showcasing.

## Anthropic / Claude family (community-observed)

No peer-reviewed Claude lexical list exists yet (2503.01659 confirms the
fingerprint but not a word set), so these are community-observed tells
([blogpros](https://blogpros.com/claude-written-content-biggest-tells/),
[Bloomberry](https://www.bloomberry.ai/research/ai-writing-patterns/claude),
[context-link](https://www.context-link.ai/blog/claude-em-dash-remover)) — mark
them lower-confidence:

- **Em-dash overuse** — the single strongest Claude punctuation tell.
- Phrase tells: `it's important to note`, `I appreciate`, `Let me…`, `Here's…`,
  `You're absolutely right`, `That said,`, `to be clear,`, `I'd be happy to`,
  `genuinely`, `nuanced`.
- Sycophantic openers: `Great question`, `You're absolutely right`.

## Cross-family structural tells

Not lexical but strong and deterministic — worth their own rules:

- **Negative parallelism:** `Not only X but also Y`, `Not X, but Y`,
  `It's not just X — it's Y`.
- **Rule of three:** three-item adjective/clause groupings, over-used.
- **Copula avoidance:** `serves as`, `stands as`, `boasts`, `features` in place of
  `is`/`has`.

## The caveat

A vale word-match cannot replicate 2503.01659's classifier and must never claim
"written by GPT/Claude." Frame family lists as: *this is elevated-register
vocabulary common to LLM output; prefer a plain, concrete term.* The lists also
decay ([2502.09606](../findings/references.md)) — keep them versioned, tagged by
family, and suppressible.
