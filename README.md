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
vale lint [flags] <path>...   Check files or directories.
vale mcp                      Start the stdio MCP server.
vale gen [flags]              Build the vocabulary rules from the wordset.
vale rules                    List the rules.
vale version                  Print the version.
```

`vale lint` accepts one or more file or directory paths. When you pass a
directory, vale walks it and checks files with a known text ending (`.md`,
`.markdown`, `.mdown`, `.mkd`, `.txt`), while it skips `.git` and `node_modules`.

### Flags for `vale lint`

| Flag | Values | Default | Meaning |
| --- | --- | --- | --- |
| `--config` | path | discovered | Path to a config file. When empty, vale searches upward from the working directory. |
| `--min-severity` | `error`, `warning`, `suggestion` | `error` | The exit-code gate. Vale fails when it finds a problem at this level or higher. |
| `--format` | `text`, `json` | `text` | Output format. |
| `--markdown` | `auto`, `on`, `off` | `auto` | Markdown mode. `auto` decides from the file ending. |
| `--strict-vocabulary` | bool | `false` | Also report unapproved words that have no direct replacement. |

### Example output

```
$ vale lint notes.txt
notes.txt
  1:48  warning     Passive voice: "was written".  [STE.PassiveVoice]
                    hint: Rewrite the sentence in the active voice. State who does the action.
  1:74  suggestion  The -ing form "using" is hard to read in an instruction.  [STE.IngForms]
                    hint: Use a simple verb form, such as the imperative.

2 problems (0 errors, 1 warnings, 1 suggestions)
```

JSON output (`--format json`) writes a `results` array; each result has a `path`
and a `findings` array. Each finding has `ruleId`, `severity`, `message`, `hint`,
`path`, `line`, `col`, `endLine`, `endCol`, and `match`. This format is stable
and machine-readable, which suits editor plugins and agents.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Clean at the gate (no finding at or above `--min-severity`). |
| `1` | One or more findings at or above the gate severity. |
| `2` | Usage error or a runtime error (bad flag, missing path, unreadable file, bad config). |

The gate defaults to `error`, so a suggestion or warning alone still exits `0`.
Raise the gate in CI (for example `--min-severity warning`) to make warnings
fail the build.

### Other subcommands

- `vale rules` prints the rule table (ID, severity, description).
- `vale gen` regenerates the vocabulary source from the wordset (see
  [Self-referential generation](#self-referential-generation)).
- `vale version` prints the build version.

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
| `STE.Vocabulary` | `suggestion` | Reports words that are not approved STE, with an approved replacement from the OpenSTE wordset. |

Every rule reports a stable ID, a severity, and a fix hint. You can disable a
rule or change its severity in the config file.

## Configuration

Vale reads an optional YAML file named `.vale-ste.yml` (or `.vale-ste.yaml`). It
searches upward from the working directory, so a file at the repository root
applies to the whole tree. Pass `--config <path>` to point at a specific file.
All fields have safe defaults, so the file is optional. See
[`.vale-ste.example.yml`](.vale-ste.example.yml) for a full example.

```yaml
# The gate for the command-line exit code: error, warning, or suggestion.
minSeverity: error

# Turn on the check for unapproved words that have no direct replacement.
strictVocabulary: false

# Word limits for the sentence-length rule.
sentence:
  procedureMax: 20
  descriptionMax: 25

# Override one rule: disable it or change its severity.
rules:
  STE.PassiveVoice:
    severity: warning
  STE.Vocabulary:
    disabled: false
```

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

### Client configuration

Add vale to an MCP client (for example Claude Code) as a server that runs the
binary with the `mcp` subcommand:

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
