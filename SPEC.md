# vale — build spec

> A single, minimal-footprint Go binary that is **both** an STE (ASD-STE100 /
> Simplified Technical English) linter **and** a self-powered MCP server. It lints
> its own docs and generates its own vocabulary rules from a vendored open wordset —
> it uses itself to build itself.

This file is the contract for the initial build. Keep it current: as you make
design decisions, record the *why* here or in `docs/decisions/`.

## Non-negotiables (from the requester)

1. **Language:** Go. **Pure Go, `CGO_ENABLED=0`** so it cross-compiles trivially.
2. **One binary, two faces:**
   - `vale lint <paths...>` — CLI linter (Simplified Technical English rules).
   - `vale mcp` — stdio MCP server exposing the same engine as MCP tools
     ("self-powered MCP"). Minimal footprint: prefer stdlib + the fewest deps that
     clear the quality bar; hand-roll stdio JSON-RPC or use one small MCP lib —
     justify the choice in `docs/decisions/`.
3. **Cross-compiles** to darwin/linux/windows × amd64/arm64 via **goreleaser**.
4. **Self-referential ("uses itself to generate itself"):**
   - `go generate ./...` regenerates the vocabulary rules from
     `third_party/openste/openste.json` (vendored, MIT — do not hand-edit generated
     rule files; regenerate). This is `vale gen` too.
   - CI lints **this repo's own `CLAUDE.md` and `AGENTS.md`** with the freshly built
     binary and fails on regressions ("uses itself").
5. **Skills:** ship Claude Code skill(s) under `skills/` that wrap the linter/MCP so
   an agent can invoke STE linting. Include a `SKILL.md` per skill.
6. **High-quality CI/CD + test suite — "gremlined, the works":**
   - `go test -race` with meaningful coverage; table-driven tests per rule.
   - **Mutation testing with [gremlins](https://github.com/go-gremlins/gremlins)**
     ("gremlined") with an efficacy threshold gate.
   - **Go native fuzzing** (`go test -fuzz`) on the tokenizer/parser and rule engine.
   - `golangci-lint run` clean (vendor a sensible `.golangci.yml`).
   - Release pipeline: goreleaser on tag → cross-compiled archives + checksums;
     optionally a Homebrew tap (the requester has `stuffbucket/homebrew-tap`).

## What to incorporate from OpenSTE (already vendored)

`third_party/openste/openste.json` (MIT, © openSTE.org) contains:
- `words[]`: `{title, plural, wordstatus: approved|unapproved, spacypos, appmeaning}`
  — 1951 words (909 approved / 1042 unapproved).
- `alternatives[]`: `{title, plural, alt_title}` — **1589 unapproved→approved pairs**
  (e.g. `abandon→stop`, `able→can`, `abnormal→unusual`). **This is the high-value
  part**: generate a Vale-style **substitution** rule from it ("word X is not STE —
  use Y"). Keep provenance + the MIT LICENSE with the generated output.

Do **not** ship the raw 1042-word "unapproved" list as a bare existence rule — on
ordinary prose it is a false-positive firehose. The curated *substitution* pairs are
the signal. (You may offer it behind an opt-in `--strict-vocabulary` flag.)

## STE rules to implement (structural; author by hand)

Each rule = its own file, table-driven tests, a stable ID, a severity, a fix hint.
- **Sentence length** ≤ 20 words (procedures) / ≤ 25 (descriptions). Configurable.
- **No contractions** (`don't`, `it's`, `can't`, …).
- **Passive voice** (be-verb + past participle) — warn.
- **`-ing` verb forms** in instructions (allow technical terms/headings) — suggestion.
- **Phrasal verbs** — small curated substitution set.
- **One instruction per sentence** / imperative for procedures — best-effort, warn.
- **Vocabulary** — generated from OpenSTE `alternatives` (suggestion level).

Severity model: `error | warning | suggestion`. Config file (native YAML is fine;
if you make it Vale-config-compatible, document why). CI should be able to gate on a
min severity (default: fail on `error`).

## Suggested layout (adapt as needed, justify deviations)

```
cmd/vale/            main.go (cobra or stdlib flag — minimal footprint)
internal/lint/       engine, tokenizer, rule interface, findings
internal/rules/      one file per rule + _test.go
internal/vocab/      generated STE vocabulary (from go generate) + generator
internal/mcp/        stdio MCP server over the engine
internal/config/     config load/merge
skills/ste-lint/     SKILL.md + wrapper
third_party/openste/ openste.json, LICENSE, PROVENANCE.txt   (already present)
docs/decisions/      ADRs
.github/workflows/   ci.yml (lint+test+race+fuzz-smoke+gremlins+self-lint), release.yml
.goreleaser.yaml     cross-compile matrix
.golangci.yml
CLAUDE.md AGENTS.md  the repo's own instructions — MUST pass `vale lint` (self-ref)
README.md LICENSE    repo is MIT
```

## Definition of done (verify before reporting complete)

- `go build ./...` and `go vet ./...` clean.
- `go test -race ./...` green with real coverage (report the %).
- `golangci-lint run` clean.
- `gremlins unleash` runs and meets the configured efficacy threshold (report it).
- At least one `go test -fuzz` target runs a short smoke without crashing.
- `vale gen` reproduces the committed vocabulary rules with no diff (idempotent).
- The built binary lints `CLAUDE.md` + `AGENTS.md` and they PASS (self-lint green) —
  i.e. write those two files IN Simplified Technical English.
- Cross-compile check: `goreleaser build --snapshot --clean` produces all targets.
- MCP server: a scripted stdio round-trip (initialize + a `lint_text` tool call)
  returns findings.
- README documents install, CLI usage, MCP usage, the skill, and the OpenSTE credit.

## Licensing / attribution

Repo license: **MIT**. Keep `third_party/openste/LICENSE` and credit OpenSTE
(MIT, © openSTE.org) in the README and in any generated vocabulary file header.
State plainly that this is an *approximation* of ASD-STE100, not certified, and that
the ASD dictionary itself is copyrighted and not bundled.
