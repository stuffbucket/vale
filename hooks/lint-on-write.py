#!/usr/bin/env python3
"""vale PostToolUse hook.

After the agent writes or edits a text file, lint it for Simplified Technical
English and slop and feed any findings back as context. This makes vale trigger
on slop itself, rather than waiting for the agent to call the MCP lint tool.

Non-blocking: it never fails the tool call. It emits additionalContext only when
there are findings, so a clean write stays silent. Requires the `vale` binary on
PATH (brew install stuffbucket/tap/vale).
"""
import json
import os
import shutil
import subprocess
import sys

LINT_EXTS = (".md", ".markdown", ".mdown", ".mkd", ".txt")


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except Exception:
        return

    path = (data.get("tool_input") or {}).get("file_path", "")
    if not path.endswith(LINT_EXTS) or not os.path.isfile(path):
        return
    if not shutil.which("vale"):
        return

    # --slop turns on the opt-in slop family; --audit never fails the caller.
    result = subprocess.run(
        ["vale", "--slop", "--audit", path],
        capture_output=True,
        text=True,
    )
    body = result.stdout.strip()
    if not body:
        return

    message = (
        "vale flagged Simplified Technical English / slop issues in "
        f"{path}. Consider tightening the writing:\n{body}"
    )
    print(
        json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PostToolUse",
                    "additionalContext": message,
                }
            }
        )
    )


if __name__ == "__main__":
    main()
