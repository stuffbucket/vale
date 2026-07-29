# Configuration, vocabulary, and suppression

## Config layers

vale merges YAML config from broad to narrow (lowest to highest precedence):
system (`$XDG_CONFIG_DIRS/vale-ste/config.yml`), user
(`$XDG_CONFIG_HOME/vale-ste/config.yml`), project (`.vale-ste.yml`),
project-local (`.vale-ste.local.yml`), the learned vocab store
(`$XDG_STATE_HOME/vale-ste/vocab.yml`), then `--config <path>`. Scalars from a
higher layer win; `vocabulary.allow`/`deny` accumulate across layers.

Copy [templates/vale-ste.yml](../templates/vale-ste.yml) to `.vale-ste.yml` to
start. Common keys:

```yaml
minSeverity: error          # exit-code gate
slop:
  enabled: false            # turn the STE.Slop* family on (or use --slop)
vocabulary:
  allow: [copilot, tauri]   # approve project terms so they stop being flagged
  deny: []                  # re-check terms
files:
  include: []               # globs; empty lints everything
  exclude: ["vendor/*"]     # exclude wins; a .valeignore adds patterns too
rules:
  STE.PassiveVoice:
    severity: warning        # or `disabled: true`
```

## Technical vocabulary

vale ships a built-in software and design term set (commit, cache, token,
component, endpoint, viewport…), so `STE.Vocabulary` does not flag them. Add your
own with `vocabulary.allow`. To approve a term for the whole session through the
MCP, call the `update_vocabulary` tool (see mcp-and-eval.md); it persists to the
learned store, which the CLI honors too.

## Inline suppression

Silence a rule on a line or region with HTML-comment directives (reusing Vale's
`off`/`on`/`= NO` and markdownlint's `disable-line` verbs):

```markdown
<!-- vale disable-next-line STE.PassiveVoice -->
This line's passive voice is ignored.

Trailing form. <!-- vale disable-line STE.Vocabulary -->

<!-- vale off -->
Everything here is ignored, for all rules.
<!-- vale on -->

<!-- vale STE.Contractions = NO -->
Only STE.Contractions is off here.
<!-- vale STE.Contractions = YES -->
```

With no rule ids, `disable-line` / `disable-next-line` suppress all rules; with
ids, only those.

## Markdown handling

vale parses Markdown to an AST and lints prose only — it skips code, raw HTML and
`<style>` CSS, link URLs, table delimiters, and YAML frontmatter. Pass
`--markdown off` to lint a file as plain text.
