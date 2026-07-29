# Research

Background research that informs vale's rules. Each topic gets its own subfolder,
following the structure of [Leonxlnx/taste-skill](https://github.com/Leonxlnx/taste-skill)
(MIT) — root causes, findings, remediation.

## Topics

### [AI Slop](ai-slop/)

Evidence on the lexical, formulaic, and structural markers of LLM-generated
prose — the "slop" — and how (and whether) a Simplified Technical English linter
can flag it. Drawn from a 16-paper arXiv sweep. Covers root causes, the empirical
findings, and the candidate `STE.Slop*` rules, with an explicit list of anti-rules
the evidence refutes.

## Scope: two halves of the same problem

Slop has a generation half and a detection half.

- **Generation-time** (do not produce slop): the [taste-skill](https://github.com/Leonxlnx/taste-skill)
  research on LLM laziness/truncation — prompt engineering, parameter tuning,
  architecture. Summarized and credited in [ai-slop/root-causes](ai-slop/root-causes/)
  and [ai-slop/remediation/design-principles.md](ai-slop/remediation/design-principles.md).
- **Detection-time** (flag slop that exists): vale's own remit. This is where the
  arXiv evidence and the candidate rules live.

vale is a detector, so the caveats matter as much as the rules: every
discriminative result in the literature is corpus-level, not per-line. See
[ai-slop/remediation/anti-rules.md](ai-slop/remediation/anti-rules.md).
