#!/bin/sh
# fix.sh — rewrite a document into Simplified Technical English with a model.
#
# It prefers a `vale` binary on PATH and falls back to
# `go run github.com/stuffbucket/vale/cmd/vale`. Every argument is passed through
# to `vale --fix`, so you can add flags and the file, for example:
#
#   ./fix.sh draft.md                       # corrected doc to stdout
#   ./fix.sh draft.md --output draft.md     # rewrite in place
#   ./fix.sh --temperature 0.2 draft.md
#
# By default it needs an LLM endpoint at http://localhost:4141 (override with
# --endpoint). It exits 0 on success, 1 on a model or write error, 2 on usage.

set -eu

if command -v vale >/dev/null 2>&1; then
	exec vale --fix "$@"
fi

if command -v go >/dev/null 2>&1; then
	exec go run github.com/stuffbucket/vale/cmd/vale --fix "$@"
fi

echo "fix.sh: neither 'vale' nor 'go' was found on PATH" >&2
exit 2
