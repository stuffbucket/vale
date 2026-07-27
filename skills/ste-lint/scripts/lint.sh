#!/bin/sh
# lint.sh — run the vale Simplified Technical English linter in JSON mode.
#
# It prefers a `vale` binary on PATH. When none is found, it falls back to
# `go run github.com/stuffbucket/vale/cmd/vale`. Every argument is passed
# through to `vale lint`, so you can add flags and paths, for example:
#
#   ./lint.sh README.md docs/
#   ./lint.sh --min-severity warning CLAUDE.md
#
# The script exits with the linter's own exit code: 0 clean at the gate,
# 1 findings at or above the gate, 2 usage or runtime error.

set -eu

if command -v vale >/dev/null 2>&1; then
	exec vale lint --format json "$@"
fi

if command -v go >/dev/null 2>&1; then
	exec go run github.com/stuffbucket/vale/cmd/vale lint --format json "$@"
fi

echo "lint.sh: neither 'vale' nor 'go' was found on PATH" >&2
exit 2
