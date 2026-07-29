# vale

A single, pure-Go binary that is **both** a Simplified Technical English
(ASD-STE100) command-line linter **and** a self-powered stdio MCP server. It
lints its own documentation and generates its own vocabulary rules from a
vendored open wordset — it uses itself to build itself.

- **CLI linter** — `vale lint <paths...>` checks text and Markdown files against
  a set of Simplified Technical English (STE) rules.
- **MCP server** — `vale mcp` exposes the same engine as Model Context Protocol
  tools over stdio, so an agent lints text through the same code the CLI uses.
- **Self-referential** — `vale gen` regenerates the vocabulary rules from a
  vendored, MIT-licensed OpenSTE wordset, and CI lints this repo's own
  `CLAUDE.md` and `AGENTS.md` with the freshly built binary.

## Disclaimer

This tool is an **approximation** of ASD-STE100 (Simplified Technical English).
It is **not certified** and it is not affiliated with or endorsed by the ASD
(AeroSpace and Defence Industries Association of Europe). The official ASD-STE100
specification and its dictionary are **copyrighted** and are **not bundled** with
this project. Vale reproduces a subset of the *structural* rules and derives a
vocabulary check from the separate, openly licensed OpenSTE wordset (see
[Credits](#credits-and-licensing)). Use it as a helpful writing aid, not as a
certificate of compliance.

## Install

With the Go toolchain:

```sh
go install github.com/stuffbucket/vale/cmd/vale@latest
```

From a release archive: cross-compiled archives and checksums are published by
[goreleaser](https://goreleaser.com/) for darwin, linux, and windows on amd64 and
arm64. Download the archive for your platform and put the `vale` binary on your
`PATH`.

With Homebrew (via the project tap):

```sh
brew install stuffbucket/tap/vale
```

From source:

```sh
git clone https://github.com/stuffbucket/vale
cd vale
go build ./cmd/vale
```

## CLI usage

```
vale [flags] <path>...        Check files or directories (the default action).
vale lint [flags] <path>...   The same, stated explicitly.
vale mcp                      Start the stdio MCP server.
vale gen [flags]              Build the vocabulary rules from the wordset.
vale rules                    List the rules.
vale version                  Print the version.
```

Linting is the default action, so `vale README.md` and `vale lint README.md` are
the same. vale accepts one or more file or directory paths. When you pass a
directory, vale walks it and checks files with a known text ending (`.md`,
`.markdown`, `.mdown`, `.mkd`, `.txt`), while it skips `.git` and `node_modules`.

### Flags for `vale lint`

| Flag | Values | Default | Meaning |
| --- | --- | --- | --- |
| `--config` | path | discovered | Path to a config file. When empty, vale searches upward from the working directory. |
| `--min-severity` | `error`, `warning`, `suggestion` | `error` | The exit-code gate. Vale fails when it finds a problem at this level or higher. |
| `--format` | `concise`, `text`, `json` | `concise` | Output format. `concise` groups by rule (token-efficient, for tools/LLMs); `text` is one verbose line per finding; `json` is machine-readable. |
| `--audit` | bool | `false` | Audit only: print findings but always exit 0 (never fail the caller). |
| `--markdown` | `auto`, `on`, `off` | `auto` | Markdown mode. `auto` decides from the file ending. |
| `--strict-vocabulary` | bool | `false` | Also report unapproved words that have no direct replacement. |
| `--slop` | bool | `false` | Enable the opt-in `STE.Slop*` rule family for this run. |
| `--color` | `auto`, `always`, `never` | `auto` | Color the severity word. `auto` colors only when stdout is a terminal and `NO_COLOR` is unset. |
| `--fix` | bool | `false` | Rewrite the file with a model so it resolves its findings; prints to stdout (see [Fixing a document](#fixing-a-document)). |

Flags may appear before, between, or after the paths — `vale README.md --format json` works.

### Fixing a document

`vale --fix <file>` lints the file, sends the original text plus the findings to a
model on the configured endpoint, and prints the corrected document to **stdout**
(so you can pipe or redirect it). Pass `--output <file>` to write to a file
instead. It fixes STE findings and the opt-in slop markers together.

```
$ vale --fix --temperature 0.2 draft.md > clean.md
$ vale --fix draft.md --output draft.md          # rewrite in place
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--model` | `model.name` config (`claude-sonnet-5`) | Model to rewrite with. |
| `--endpoint` | `model.endpoint` config (`http://localhost:4141`) | OpenAI-compatible base URL. |
| `--temperature` | `model.temperature` config / server default | Sampling temperature. |
| `--output` | stdout | Write the corrected document here. |
| `--max-tokens` | `2048` | Token budget for the rewrite. |

### Example output

The default `concise` format groups findings by rule, prints the rule ID and a
shared hint once, then one compact `path:line:col match` line per finding. It
drops the repeated per-line rule ID and message that agents and LLMs would
otherwise pay for in tokens, while keeping every location actionable.

```
$ vale notes.txt
STE.PassiveVoice · warning · 1
  ↳ Rewrite the sentence in the active voice. State who does the action.
  notes.txt:1:48  was written

STE.IngForms · suggestion · 1
  ↳ Use a simple verb form, such as the imperative.
  notes.txt:1:74  using
```

Use `--format text` for one self-contained clickable line per finding
(`path:line:col: severity: message [RuleID] hint: …`), which suits editors and
CI problem matchers. Findings go to stdout; the summary count goes to stderr, so
a pipe over stdout gets nothing but findings. Output adapts to the caller: at an
interactive terminal the severity word is colored; piped, redirected, or under
CI it is plain text.

JSON output (`--format json`) writes a `results` array; each result has a `path`
and a `findings` array. Each finding has `ruleId`, `severity`, `message`, `hint`,
`path`, `line`, `col`, `endLine`, `endCol`, and `match`. This format is stable
and machine-readable, which suits editor plugins and agents.

### Exit codes

The codes follow the convention agentic harnesses expect: `0` success, `1`
lint findings, `2` the tool itself failed.

| Code | Meaning |
| --- | --- |
| `0` | Clean at the gate (no finding at or above `--min-severity`), or `--audit`. |
| `1` | One or more findings at or above the gate severity. |
| `2` | Usage error or a runtime error (bad flag, missing path, unreadable file, bad config). |

Pass `--audit` to always exit `0` — the findings still print, but the run never
fails the caller. This suits an agent that wants to collect findings without a
non-zero exit tripping its command-failure handling.

The gate defaults to `error`, so a suggestion or warning alone still exits `0`.
Raise the gate in CI (for example `--min-severity warning`) to make warnings
fail the build.

### Other subcommands

- `vale rules` prints the rule table (ID, severity, description).
- `vale gen` regenerates the vocabulary source from the wordset (see
  [Self-referential generation](#self-referential-generation)).
- `vale eval` measures slop across an LLM endpoint's models (see
  [Evaluating models](#evaluating-models)).
- `vale version` prints the build provenance: the semantic version, the source
  branch and commit, and the build timestamp (for example
  `vale 0.2.0 (branch main, commit 0d0c6fee13f6, built 2026-07-29T10:16:21Z)`).

## Evaluating models

`vale eval` turns the [slop research](research/ai-slop/) into a measurement. It
drives a configured OpenAI-compatible endpoint (by default `http://localhost:4141`),
prompts each model, lints every reply with the `STE.Slop*` rules, and reports
slop density per model and per family. At an interactive terminal it shows a live
progress bar on stderr while the requests run. Model families come from the endpoint's
`/v1/models` `owned_by` field; models the endpoint only serves through the
Responses API are reached automatically via a `/v1/responses` fallback.

```
$ vale eval --models "claude-sonnet-5,gemini-2.5-pro,gpt-5.5,grok-4.5"
family slop density (STE.Slop findings per 100 words), most slop first:
  xAI          slop   0.93  all  17.59  (1 models, 324 words)
  OpenAI       slop   0.67  all  19.06  (1 models, 299 words)
  Anthropic    slop   0.65  all  17.63  (1 models, 465 words)
  Google       slop   0.43  all  15.00  (1 models, 460 words)
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--endpoint` | `http://localhost:4141` | OpenAI-compatible base URL. |
| `--models` | discover all | Comma-separated model ids; empty discovers every model from `/v1/models`. |
| `--prompts` | built-in set | File with one prompt per line. |
| `--max-tokens` | `400` | Max tokens per completion. |
| `--temperature` | server default | Sampling temperature (or set `model.temperature` in config). |
| `--concurrency` | `4` | Parallel requests. |
| `--format` | `text` | `text` or `json`. |
| `--api-key` | `$OPENAI_API_KEY` | Bearer token, if the endpoint needs one. |

## Rules

`vale rules` lists every rule with its default severity:

| ID | Default severity | What it checks |
| --- | --- | --- |
| `STE.SentenceLength` | `error` | Keeps sentences short: procedures (list items) to 20 words, descriptions to 25 words. Both limits are configurable. |
| `STE.Contractions` | `error` | Flags contractions such as `don't`, `it's`, and `can't`. Write the full words. |
| `STE.PassiveVoice` | `warning` | Detects the passive voice (a be-verb plus a past participle). Prefer the active voice. |
| `STE.IngForms` | `suggestion` | Flags `-ing` verb forms in instructions, which are hard to read. |
| `STE.PhrasalVerbs` | `warning` | Flags phrasal verbs from a curated set; prefer one clear verb. |
| `STE.OneInstruction` | `warning` | Asks for one instruction in one sentence. |
| `STE.Vocabulary` | `suggestion` | Reports words that are not approved STE, with an approved replacement from the OpenSTE wordset. Approved technical terms are skipped (see [Technical terms](#technical-terms)). |

Every rule reports a stable ID, a severity, and a fix hint. You can disable a
rule or change its severity in the config file.

### Technical terms

ASD-STE100 does not expect its core dictionary to name a domain; it lets a
project approve **Technical Names** (nouns) and **Technical Verbs** for its own
field. vale ships a built-in set for the software development lifecycle and for
design — `commit`, `branch`, `cache`, `token`, `endpoint`, `component`,
`viewport`, `design token`, and many more — so `STE.Vocabulary` does not flag
them or suggest an aerospace-sense replacement (for example `file` → `remove`).

Tune the set in the `vocabulary` block of the config:

```yaml
vocabulary:
  builtinTechnicalTerms: true # the built-in software and design set (default on)
  allow: [copilot, tauri, sidecar] # project terms, single words or phrases
  deny: [] # remove terms from the approved set to check them again
```

## Configuration

vale resolves configuration from several files, broad to narrow, the way
XDG-aware tools do. Lowest to highest precedence:

| Layer | Location |
| --- | --- |
| System | `$XDG_CONFIG_DIRS/vale-ste/config.yml` (default `/etc/xdg`) |
| User | `$XDG_CONFIG_HOME/vale-ste/config.yml` (default `~/.config`) |
| Project | the nearest `.vale-ste.yml`, searching up from the working directory |
| Project-local | the nearest `.vale-ste.local.yml` (personal, git-ignore it) |
| Explicit | `--config <path>`, when given |

A scalar from a higher layer replaces a lower one (a project `minSeverity` beats
the user default). `vocabulary.allow` and `vocabulary.deny` **accumulate** across
layers, so a personal allow list in `~/.config/vale-ste/config.yml` adds to the
project's list rather than replacing it. Rule settings merge per rule id. A
missing file at any layer is skipped, and all fields have safe defaults, so no
config is required.

```yaml
# The gate for the command-line exit code: error, warning, or suggestion.
minSeverity: error

# Turn on the check for unapproved words that have no direct replacement.
strictVocabulary: false

# Word limits for the sentence-length rule.
sentence:
  procedureMax: 20
  descriptionMax: 25

# Scope which paths a directory walk considers (globs; exclude wins over
# include). A .valeignore file adds exclude patterns too. Explicitly named file
# arguments are always linted.
files:
  include: [] # e.g. ["docs/**/*.md"] — when set, only matches are linted
  exclude: ["vendor/*", "**/CHANGELOG.md"]

# Override one rule: disable it or change its severity.
rules:
  STE.PassiveVoice:
    severity: warning
  STE.Vocabulary:
    disabled: false
```

## Suppressing rules inline

Silence a rule on a line or region with HTML-comment directives. The notation
reuses Vale's own `off`/`on`/`= NO` region syntax and markdownlint's
`disable-line` / `disable-next-line` verbs — nothing new to learn.

```markdown
This line is checked.

<!-- vale disable-next-line STE.PassiveVoice -->
This line's passive voice is ignored.

Trailing form works too. <!-- vale disable-line STE.Vocabulary -->

<!-- vale off -->
Everything here is ignored, for every rule,
until the matching on directive.
<!-- vale on -->

<!-- vale STE.Contractions = NO -->
Only STE.Contractions is off in this region.
<!-- vale STE.Contractions = YES -->
```

With no rule ids, a `disable-line` / `disable-next-line` directive suppresses
every rule; with one or more space-separated ids, only those. Directives are
HTML comments, so they never produce findings themselves. See
[ADR 0007](docs/decisions/0007-inline-suppression-directives.md).

## How Markdown is parsed

vale parses Markdown to a real AST ([goldmark](https://github.com/yuin/goldmark),
CommonMark + GFM) and lints **prose nodes only** — paragraphs, headings, list
items, and table cells. It skips structure by construction: fenced and inline
code, raw HTML and `<style>`/`<script>` CSS, link and image URLs, table
delimiters, and YAML frontmatter. So HTML/CSS embedded in a document does not
produce spurious findings, and positions stay `path:line:col` accurate. Pass
`--markdown off` to lint a file as plain text instead.

Fields:

- `minSeverity` — the exit-code gate (`error`, `warning`, or `suggestion`).
- `strictVocabulary` — when `true`, also report unapproved words with no
  replacement. Off by default, because it makes many findings on ordinary prose.
- `sentence.procedureMax` / `sentence.descriptionMax` — the word limits for the
  sentence-length rule.
- `rules.<ID>.disabled` — set `true` to turn a rule off.
- `rules.<ID>.severity` — override a rule's severity.

## MCP usage

`vale mcp` starts a stdio MCP server. It speaks JSON-RPC 2.0 with one JSON
message on each line (newline-delimited), and it reports MCP protocol version
`2024-11-05`. It exposes two tools:

- **`lint_text`** — check a string against the STE rules.
  Arguments: `text` (required), `filename` (optional; a `.md` name turns on
  Markdown mode), `markdown` (optional boolean; overrides the file name), and
  `minSeverity` (optional; `error`, `warning`, or `suggestion`).
- **`list_rules`** — list the rules the linter uses.
- **`fix_text`** — rewrite text with a model so it resolves its findings; returns
  the corrected document. Arguments: `text` (required), `filename`, `model`
  (defaults to `model.name`), `temperature` (optional override), and `maxTokens`.
- **`update_vocabulary`** — learn project vocabulary for the session. Arguments:
  `allow` (terms to approve) and `deny` (terms to re-check). It persists terms to
  the vocab store and rebuilds the linter, so later `lint_text` and `fix_text`
  calls honor them — the server adapts to the session.

The server runs in session mode. Learned terms persist to a vocab store —
`$XDG_STATE_HOME/vale-ste/vocab.yml` (`~/.local/state/vale-ste/vocab.yml`) by
default, the XDG location for state that survives restarts. The CLI and the hook
read it too via config layering, so learning is honored everywhere. Scope the
store to one session with `vale mcp --vocab-store /run/user/…/vale-<session>.yml`,
or leave the default to share learning across sessions and projects.

### Client configuration

The repository ships a ready-to-use MCP definition and a Claude Code plugin.

**As a Claude Code plugin** (recommended). The repo is a plugin marketplace, so
it installs the MCP server, the `ste-lint` skill, and a lint-on-write hook
together:

```shell
/plugin marketplace add stuffbucket/vale
/plugin install vale@stuffbucket
/reload-plugins
```

The plugin's [`PostToolUse` hook](hooks/lint-on-write.py) lints every text file
the agent writes or edits (`--slop --audit`) and feeds any findings back as
context. So vale triggers on slop itself, rather than waiting for the agent to
call a tool — while the MCP definition stays a single tight server.

Install the `vale` binary first (the plugin launches it): `brew install
stuffbucket/tap/vale`. The plugin's server definition lives in
[`.mcp.json`](.mcp.json); its manifest is
[`.claude-plugin/plugin.json`](.claude-plugin/plugin.json) and the marketplace is
[`.claude-plugin/marketplace.json`](.claude-plugin/marketplace.json).

**Directly, without the plugin.** Add vale to any MCP client (Claude Code,
Claude Desktop) as a server that runs the binary with the `mcp` subcommand — the
same JSON the repo's `.mcp.json` contains:

```json
{
  "mcpServers": {
    "vale": {
      "command": "vale",
      "args": ["mcp"]
    }
  }
}
```

### Raw stdio example

Each line below is one JSON-RPC message sent to the server's standard input; the
server writes one JSON response line per request. Notifications (no `id`) get no
response.

```
--> {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"demo","version":"1"}}}
<-- {"jsonrpc":"2.0","id":1,"result":{"capabilities":{"tools":{}},"protocolVersion":"2024-11-05","serverInfo":{"name":"vale-ste","version":"dev"}}}
--> {"jsonrpc":"2.0","method":"notifications/initialized"}
--> {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lint_text","arguments":{"text":"The system was started by the operator.","minSeverity":"warning"}}}
<-- {"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"1 findings (0 errors, 1 warnings, 0 suggestions)\n{ ... findings JSON ... }"}],"isError":false}}
```

The tool result `content` is a text block. It starts with a one-line summary and
then a JSON block that holds the `findings` array and a `summary` count by
severity.

## Self-referential generation

The vocabulary rule is generated, not hand-written. `vale gen` reads
`third_party/openste/openste.json` and writes
`internal/vocab/substitutions_gen.go`. The same work runs through `go generate`:

```sh
vale gen            # or: go run ./cmd/vale gen
go generate ./...   # runs the //go:generate directive in internal/vocab
```

Do **not** hand-edit `internal/vocab/substitutions_gen.go`. Change the wordset or
the generator and run `vale gen` again. The generator is **idempotent**: running
it on an unchanged wordset produces no diff, which lets CI verify the committed
file is current.

## Self-lint

This repository's own `CLAUDE.md` and `AGENTS.md` are written in Simplified
Technical English. CI lints them with the freshly built binary and fails on a new
error, so the tool holds its own documentation to the standard it enforces:

```sh
vale lint CLAUDE.md AGENTS.md
```

## Skill

A Claude Code skill wraps the linter so an agent can lint or rewrite text into
Simplified Technical English. See
[`skills/ste-lint/SKILL.md`](skills/ste-lint/SKILL.md). A small shell wrapper
lives at [`skills/ste-lint/scripts/lint.sh`](skills/ste-lint/scripts/lint.sh).

## Development

```sh
go build ./...            # build every package
go test -race ./...       # run the tests with the race detector
golangci-lint run         # lint the Go source
```

The project ships a `Makefile` with convenience targets:

| Target | Purpose |
| --- | --- |
| `build` | Build the binary. |
| `test` | Run the tests with the race detector. |
| `lint` | Run `golangci-lint`. |
| `cover` | Run the tests and report coverage. |
| `gen` | Regenerate the vocabulary source (`vale gen`). |
| `self-lint` | Lint `CLAUDE.md` and `AGENTS.md` with the built binary. |
| `snapshot` | Cross-compile a goreleaser snapshot. |
| `mutation` | Run mutation testing (gremlins). |

## Architecture decisions

Design decisions are recorded as ADRs under
[`docs/decisions/`](docs/decisions/). See the
[index](docs/decisions/README.md).

## Credits and licensing

This repository is licensed under the [MIT License](LICENSE), © stuffbucket.

The vocabulary rule is derived from the **OpenSTE wordset** (openste, v1.01),
which is licensed under the MIT License, © openSTE.org. The wordset and its
license and provenance are vendored under
[`third_party/openste/`](third_party/openste/) — see
[`third_party/openste/LICENSE`](third_party/openste/LICENSE) and
[`third_party/openste/PROVENANCE.txt`](third_party/openste/PROVENANCE.txt). The
generated vocabulary file keeps this attribution in its header.

As stated in the [Disclaimer](#disclaimer), vale is an approximation of
ASD-STE100. It is not certified, and the ASD-STE100 dictionary itself is
copyrighted and is not bundled.
