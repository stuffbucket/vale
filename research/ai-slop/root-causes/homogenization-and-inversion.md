# Homogenization and Inversion

This is the most important root-cause page for a Simplified Technical English
linter, because it is where naive slop detection goes wrong.

## The homogenization signatures

At the corpus level, LLM text is measurably more uniform than human text:

- **Lower lexical diversity.** On matched tasks, GPT-3.5 used ~25-50% fewer
  distinct lemmas and a lower Root Type-Token Ratio than humans
  ([2308.07462](../findings/references.md)).
- **Shorter, simpler sentences** and reduced syntactic complexity
  ([2505.12218](../findings/references.md), [2302.00937](../findings/references.md)).
- **Dense articles/prepositions, suppressed first person**
  ([2401.16587](../findings/references.md)).
- **Higher nominalization density** — nouns derived from verbs via `-tion`,
  `-ment`, `-ance`, `-ity` (`utilization`, `implementation`, `consideration`) —
  and a hedge-light, flat stance ([2304.14276](../findings/references.md)).

## The inversion trap

Here is the problem: **well-formed STE deliberately produces most of these same
signatures.** Controlled vocabulary lowers lexical diversity. The sentence-length
rule produces short, simple sentences. Technical spec register is
nominalization-heavy and hedge-light. A document that perfectly obeys STE sits
squarely inside the "looks like AI" region for every one of these metrics.

So these signatures **cannot separate good controlled-language writing from AI
output**. Built as rules, they would systematically false-positive on exactly the
prose vale exists to encourage. They are non-signals or inverse-signals for this
tool and are documented as anti-rules, not rules
([anti-rules.md](../remediation/anti-rules.md)).

## The one that survives inversion

Nominalization is the exception. LLM text runs high on it **and** STE independently
penalizes it (STE wants the direct verb: `configure`, not `perform the
configuration`). Because the tool's goal and the slop signal point the same way
here, a nominalization-density metric is defensible — as an editorial nudge toward
verbs, exempting the technical-terms allowlist, never an autofix
([candidate-rules.md](../remediation/candidate-rules.md),
`STE.NominalizationDensity`).
