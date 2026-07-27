# 3. YAML configuration via gopkg.in/yaml.v3

- Status: Accepted

## Context

Vale needs an optional project configuration file. Users must be able to set the
exit-code gate, turn strict vocabulary on or off, adjust the sentence-length
limits, and disable or re-severity individual rules. The format should be
familiar to people who configure linters and other developer tools, and it must
stay pure Go so the binary keeps its `CGO_ENABLED=0` cross-compile property (see
[ADR 0001](0001-single-pure-go-binary.md)).

Two candidates stood out:

- **YAML**, which is the common configuration format for linters and static
  analysis tools, at the cost of one third-party dependency.
- **JSON** from the standard library, which needs no dependency but is less
  pleasant to hand-write (no comments, strict punctuation) and is unusual for a
  linter configuration file.

## Decision

Use a YAML configuration file, `.vale-ste.yml` (or `.vale-ste.yaml`), parsed with
`gopkg.in/yaml.v3`. This is the project's **single** third-party dependency.
`gopkg.in/yaml.v3` is pure Go, so it does not compromise cross-compilation.

The loader (`internal/config`) discovers the file by walking up from the working
directory, so a file at the repository root applies to the whole tree. All fields
have safe defaults, so the file is optional. A `--config` flag points at an
explicit file.

The stdlib-only JSON alternative was rejected: the small dependency buys a much
better authoring experience (comments, forgiving syntax) and matches what users
expect from a linter.

## Consequences

- The configuration reads and writes like other linter configs, with comments
  and readable structure.
- The project depends on exactly one external module, which is pure Go and
  widely used; the dependency surface stays auditable.
- The binary keeps its static, cross-compiling build.
- If the single dependency ever becomes a problem, the config layer is small and
  isolated, so a move to stdlib JSON would be contained.
