# Anti-Rules

Rules the evidence tells us **not** to build. Documented so a plausible-sounding
intuition does not get implemented later. Each would produce systematic false
positives, most of them on exactly the controlled-language prose vale exists to
encourage.

## Do not flag connective overuse

The popular belief that AI overuses `however`, `moreover`, and `furthermore` is
**contradicted** by the evidence: ChatGPT argumentative essays used *fewer*
discourse and stance connectives than humans ([2304.14276]). A connective-overuse
rule would flag well-structured human writing. Do not build it.

## Do not flag the STE-inverting signatures

Several corpus-level "AI-like" signatures **invert** for Simplified Technical
English — a compliant STE document scores "AI" on all of them. Building any of
these as a rule would false-positive on the tool's own ideal output (see
[root-causes/homogenization-and-inversion.md](../root-causes/homogenization-and-inversion.md)):

- **Low lexical diversity / low type-token ratio** ([2308.07462]) — STE's
  controlled vocabulary is intentionally low-diversity.
- **Short, simple sentences / low syntactic complexity** ([2505.12218],
  [2302.00937]) — STE's sentence-length rule *produces* these.
- **Dense articles/prepositions, suppressed first person** ([2401.16587]) — normal
  for impersonal technical register.
- **Hedge-light, flat stance** ([2304.14276]) — STE wants direct statements.

## Do not treat length as a defect

Verbosity and length are confounds, not defects ([2404.04475], [2411.07858]).
Length-based flags misfire on legitimately long reference docs and on correct STE
lists and formatting. The QA-tuned `>3-token` verbosity detector does not transfer
to prose. Only *relative/comparative* length signals are defensible, and even
those are weak.

## Do not use whole-corpus richness metrics on single documents

Lexical-richness metrics (RTTR, MTLD, Herdan) need thousands of tokens to
stabilize, are undefined on short passages, and depend on genre/token-count
matching a linter rarely has ([2308.07462], authors call the results "very
preliminary"). Not suitable at typical document sizes.

## Do not claim AI authorship

The overarching guardrail. Every discriminative result in the literature is
corpus/population-level; multiple papers explicitly disclaim per-document
inference and one calls text-by-text detection "perhaps impossible" ([2502.09606]).
A vale rule fires per line, so **no single-occurrence "this is AI" verdict is
supported by this evidence.** Every slop rule is a plain-language / style nudge,
not an authorship claim.
