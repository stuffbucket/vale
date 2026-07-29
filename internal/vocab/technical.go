// Package vocab — technical terms.
//
// ASD-STE100 does not expect its core dictionary to name a domain. The standard
// permits a project to add Technical Names (nouns) and Technical Verbs so that
// domain vocabulary is legitimately approved. TechnicalTerms is vale's built-in
// set for the software development lifecycle and for design work. It is
// hand-curated (not generated from the wordset) and holds only true domain
// vocabulary — not general-English words that STE flags for clarity (for
// example "as", "over", "both"), which stay in force.
//
// The vocabulary rule treats every term here as approved. A project extends or
// overrides the set with vocabulary.allow and vocabulary.deny in its config.
package vocab

// TechnicalTerms lists the built-in approved Technical Names and Technical Verbs
// for software and design. Entries are lower case; multiword entries are matched
// as phrases.
var TechnicalTerms = []string{
	// Version control and the development lifecycle.
	"commit", "branch", "merge", "rebase", "checkout", "clone", "fork", "push",
	"pull", "stash", "tag", "diff", "blame", "hash", "upstream", "origin",
	"remote", "repository", "repo", "changelog", "backport", "cherry-pick",
	"pull request", "commit hash", "merge conflict", "source control",

	// Build, CI/CD, and release.
	"build", "compile", "transpile", "bundle", "minify", "lint", "linter",
	"deploy", "deployment", "release", "rollback", "pipeline", "artifact",
	"runner", "workflow", "job", "cron", "provision", "bootstrap", "scaffold",
	"semver", "changeset",

	// Code and language constructs.
	"code", "codebase", "source", "function", "method", "class", "module",
	"package", "namespace", "variable", "constant", "parameter", "argument",
	"interface", "struct", "enum", "boolean", "string", "integer", "float",
	"array", "slice", "map", "object", "pointer", "reference", "callback",
	"closure", "iterator", "generic", "annotation", "decorator", "macro",
	"async", "await", "thread", "mutex", "goroutine", "coroutine", "recursion",
	"refactor", "deprecate", "instantiate", "serialize", "deserialize",
	"parse", "tokenize", "lint",

	// APIs, services, and the web.
	"api", "endpoint", "request", "response", "payload", "schema", "query",
	"mutation", "middleware", "handler", "router", "route", "proxy", "webhook",
	"rate limit", "throttle", "poll", "stream", "socket", "port", "host",
	"header", "status code", "url", "uri", "http", "https", "rest", "graphql",
	"grpc", "json", "yaml", "xml", "html", "css", "dom", "markdown",

	// State, data, and storage.
	"cache", "token", "session", "cookie", "checksum", "config", "flag",
	"environment", "runtime", "dependency", "framework", "library", "sdk",
	"cli", "binary", "executable", "database", "table", "row", "column",
	"index", "schema", "migration", "seed", "buffer", "queue", "log", "metric",
	"snapshot", "backup", "file", "directory", "path", "symlink", "blob",

	// Infrastructure and operations.
	"server", "client", "container", "image", "volume", "cluster", "node",
	"pod", "load balancer", "gateway", "firewall", "sidecar", "daemon",
	"process", "kernel", "sandbox",

	// Testing and quality.
	"test", "mock", "stub", "fixture", "assertion", "assert", "coverage",
	"fuzz", "benchmark", "regression", "flaky", "verify", "validate",
	"lint", "profiler",

	// Common technical verbs that STE flags but that carry the software sense.
	"fix", "register", "toggle", "render", "mount", "unmount", "dispatch",
	"emit", "subscribe", "publish", "sync", "index", "hash", "escape",

	// Design and UI.
	"component", "layout", "grid", "palette", "typography", "typeface",
	"font", "glyph", "kerning", "leading", "accent", "contrast", "theme",
	"token", "viewport", "breakpoint", "margin", "padding", "gutter",
	"gradient", "opacity", "hue", "saturation", "luminance", "swatch",
	"wireframe", "mockup", "prototype", "affordance", "hierarchy", "alignment",
	"whitespace", "icon", "asset", "spec", "design token", "style guide",
	"color ramp", "focus ring",
}
