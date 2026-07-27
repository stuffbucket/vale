# 4. Self-referential vocabulary generation from the OpenSTE wordset

- Status: Accepted

## Context

The Simplified Technical English vocabulary check needs a large list of
unapproved words and their approved replacements. Maintaining that list by hand
would be error-prone and would drift from any upstream source.

The official ASD-STE100 dictionary is copyrighted and cannot be bundled. A
separate, openly licensed source exists: the OpenSTE wordset (MIT, © openSTE.org),
vendored under `third_party/openste/`. Its `alternatives[]` array holds
unapproved-to-approved pairs (for example `abandon → stop`, `able → can`). This
is the high-value signal. The wordset also holds a bare list of unapproved words
with no replacement; on ordinary prose that list is a false-positive firehose.

## Decision

Generate the vocabulary rule from the vendored wordset instead of writing it by
hand. The generator (`internal/vocab/generator`) reads
`third_party/openste/openste.json` and writes
`internal/vocab/substitutions_gen.go`.

- The generated file is **checked in**, so a plain `go build` needs no generation
  step, and code review can see the data.
- Generation is **idempotent**: running it on an unchanged wordset produces no
  diff. CI regenerates and checks that git shows no change, which proves the
  committed file is current.
- `vale gen` and `go generate ./...` do the same work; a `//go:generate`
  directive in `internal/vocab` invokes `vale gen`.
- The generated header keeps the OpenSTE attribution and the ASD-STE100
  approximation disclaimer.
- By default, only the **substitution pairs** (the signal) drive the vocabulary
  rule. The bare unapproved-word list is behind the opt-in `--strict-vocabulary`
  flag, to avoid a flood of low-value findings on normal text.

## Consequences

- The vocabulary stays in sync with the vendored source; a wordset update is one
  regeneration away.
- The tool "uses itself to build itself": the `vale` binary generates part of its
  own rule set.
- Contributors must not hand-edit the generated file; they change the wordset or
  the generator and run `vale gen`. The idempotence check in CI enforces this.
- The default vocabulary check stays high-signal; strict mode remains available
  for users who want full coverage.
