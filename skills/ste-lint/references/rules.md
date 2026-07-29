# Rules and the rewrite checklist

vale reports a stable rule ID, a severity, and a fix `hint` for every finding.
Fix errors first, then warnings, then suggestions, and re-run until clean.

## Core STE rules (always on)

| ID | Severity | What it flags |
| --- | --- | --- |
| `STE.SentenceLength` | error | Sentences over the limit: 20 words in procedures (list items), 25 in descriptions. |
| `STE.Contractions` | error | Contractions (`don't`, `it's`, `can't`). Write the full words. |
| `STE.PassiveVoice` | warning | A be-verb plus a past participle. Use the active voice. |
| `STE.IngForms` | suggestion | `-ing` verb forms in instructions. Use the plain imperative. |
| `STE.PhrasalVerbs` | warning | Phrasal verbs. Use one clear verb. |
| `STE.OneInstruction` | warning | More than one instruction in a sentence. |
| `STE.Vocabulary` | suggestion | Words that are not approved STE; the `hint` gives the replacement. |

## Slop rules (opt-in, add `--slop`)

Advisory markers of LLM-generated "slop." Off by default; low-baseline signals,
not per-line verdicts.

- `STE.SlopVocabulary` — spike words (delve, intricate, realm, showcasing…).
- `STE.SlopImpersonalHedge` — "it could be argued that…" padding.
- `STE.SlopNegativeParallelism` — "not only X but also Y".
- `STE.SlopRestatement` — "in other words", "to reiterate" (reinvented points).
- `STE.SlopEvaluative` — clustered praise adjectives.
- `STE.SlopHedgeDensity` — stacked hedges in one sentence.
- `STE.SlopNominalization` — verbs frozen into nouns (utilization, implementation).

## Rewrite checklist

1. **Expand contractions** — `do not`, `it is`, `cannot`.
2. **Shorten sentences** — split one long sentence into two short ones; keep
   steps to 20 words or fewer.
3. **Use the active voice** — "close the valve", not "the valve was closed".
4. **One instruction per sentence** — one action per step, imperative for
   procedures.
5. **Drop `-ing` forms** in instructions — use the plain imperative.
6. **Replace phrasal verbs** with one clear verb.
7. **Use the approved word** from each `STE.Vocabulary` finding's `hint`.

Keep the technical meaning exact. After a pass, run `vale <path> --format json`
and confirm the findings are gone at the gate you care about.

## JSON finding shape

Each finding: `ruleId`, `severity`, `message`, `hint`, `line`, `col`, `endLine`,
`endCol`, `match`. The concise text format groups these by rule and drops the
per-line repetition; use JSON when you need exact spans to edit.
