# Repetition and Reinvention

The pattern you feel when a model says the same thing three ways, restates the
prompt, or "reinvents" a concept it already covered. This is one of the most
lint-able slop dimensions, because unlike lexical spikes it has deterministic,
surface-visible proxies.

## Why models repeat

Repetition is a well-studied failure mode of neural text generation, not a random
glitch:

- **Degeneration under likelihood decoding** — maximizing likelihood produces text
  that is "bland and strangely repetitive"
  ([1904.09751](../findings/references.md), Holtzman et al.).
- **Self-reinforcement** — once a model repeats a sentence, the probability of
  repeating it again *increases*; repetition feeds on itself
  ([2206.02369](../findings/references.md), "Learning to Break the Loop"). This is
  precisely the "over and over" effect.
- **A structural cause** — the "high inflow problem": certain tokens accumulate
  probability mass, quantified by an Average Repetition Probability
  ([2012.14660](../findings/references.md)).
- **A mechanistic locus** — specific "repetition neurons" drive the behavior
  ([2410.13497](../findings/references.md)).

Instruction-tuned models rarely loop verbatim (RLHF suppresses that), so the slop
form is **near-repetition and semantic reinvention**: the same idea rephrased,
re-summarized, or reintroduced in a new section.

## How to lint it

Three tiers, cheapest and most deterministic first.

### 1. Restatement discourse markers (existence — deterministic, low FP)
Phrases that explicitly announce a restatement: `in other words`, `to put it
another way`, `simply put`, `essentially,`, `that is to say`, `to reiterate`,
`as (previously) mentioned`, `as noted above`, `again,`, `in essence`. A cluster
of these is a strong "reinventing the same point" tell, and each has a clean
rewrite (delete the restatement or merge it). Candidate rule
`STE.RestatementMarkers`.

### 2. N-gram / phrase repetition (occurrence — standard metrics)
The generation literature's repetition metrics translate directly:

- **rep-n** — fraction of repeated n-grams (n = 3-4): `1 − distinct_ngrams/total`.
- **distinct-n** — ratio of unique n-grams to total; low distinct-n flags a
  repetitive passage.
- Flag when rep-n over a document (or a rolling window) exceeds a conservative
  threshold, exempting legitimate repeated technical terms via the allowlist.
  Candidate rule `STE.PhraseRepetition`.

### 3. Sentence-level near-duplication (metric — advisory)
"Reinvention" is two sentences that say the same thing in different words. Cheap
lexical proxy without embeddings: **shingling + Jaccard similarity** over
content-word sets per sentence pair; flag pairs above a high threshold (e.g.
Jaccard > 0.6 on lemmatized content words). This catches re-summarized paragraphs
and duplicated section intros. Self-BLEU is the corpus-level analogue. Keep it
advisory — genuine parallel structure (a spec listing similar requirements) will
score high, so this needs human review and must never autofix. Candidate rule
`STE.RedundantRestatement`.

## Boundaries

- **Semantic** reinvention (same meaning, no lexical overlap) needs embeddings,
  which a pure-Go, dependency-light linter should not take on by default; the
  lexical proxies above catch the common case. Note the gap rather than hide it.
- STE deliberately repeats **approved terms** (consistency is a virtue in
  controlled language), so any repetition rule must exempt the technical-terms
  allowlist and target repeated *phrases and sentences*, not repeated nouns.
