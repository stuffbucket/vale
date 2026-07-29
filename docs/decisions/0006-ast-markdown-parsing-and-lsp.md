# 0006. AST-based Markdown parsing, path scoping, and an LSP surface

Status: Proposed

## Context

vale's current Markdown handling is a hand-rolled, line-based masker
(`internal/lint/tokenizer.go`): it blanks fenced code, inline code, and link
destinations line by line. This cannot reliably separate **content from
structure**. Known gaps:

- Raw HTML (`<div class="card">`, `<span style="…">`) and `<style>` CSS blocks are
  not stripped — their tokens leak into the prose stream and produce noise.
- The masker's own comment claims it strips autolinks `<…>`, but the code does
  not (a latent bug).
- Tables, block quotes, and YAML frontmatter are not classified, so cell
  delimiters and metadata are linted as prose.

The right tool is a real Markdown AST that classifies every block, so the linter
reasons over prose nodes only.

## Decision

1. **Parse to an AST instead of masking lines.** The recommendation to use
   MDAST/Remark is sound in intent, but those are JavaScript/Node libraries.
   vale is a single **pure-Go, `CGO_ENABLED=0`** binary; a Node dependency would
   break that. The Go-native equivalent is **[goldmark](https://github.com/yuin/goldmark)**
   — a CommonMark-compliant, pure-Go AST parser (the engine behind Hugo), with
   GFM table and frontmatter extensions. It gives the same block-accurate lexical
   analysis MDAST/Remark would, with no runtime outside the binary. This raises
   the dependency count from one (`yaml.v3`) to two; the accuracy gain justifies
   it, and it is recorded here rather than taken silently (see ADR 0002).

2. **Walk the AST, lint prose text nodes only.** Extract text from paragraph,
   heading, list-item, table-cell, and emphasis nodes; skip code blocks/spans,
   HTML blocks/inline, `<style>`/`<script>`, link/image URLs, and frontmatter.
   Preserve source line/column via goldmark segments so findings stay
   `path:line:col` accurate.

3. **Path scoping in the layered config.** Add `files.include` / `files.exclude`
   glob settings (and honor a `.valeignore`), resolved through the existing
   config layering, so a project controls which paths are considered.

4. **An LSP surface (staged follow-up).** vale already exposes an MCP server over
   the same engine; a Language Server Protocol mode would give editors live
   diagnostics and reuse the AST block classification. Larger surface, so it lands
   after 1-3.

## Consequences

- Fixes the HTML/CSS/table/frontmatter noise and the autolink bug by construction.
- Second dependency (goldmark), pure Go, no cgo — cross-compilation unaffected.
- Block classification becomes reusable by the MCP tools and the future LSP.
- Staging: (1) goldmark prose extraction, (2) path scoping, (3) LSP.
