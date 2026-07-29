package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/linter"
	"github.com/stuffbucket/vale/internal/report"
)

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	w.Close()
	data, _ := io.ReadAll(r)
	return string(data)
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"no args", nil, 2},
		{"nonexistent path routes to lint", []string{"bogus"}, 2},
		{"version", []string{"version"}, 0},
		{"version flag", []string{"--version"}, 0},
		{"short version", []string{"-v"}, 0},
		{"help", []string{"help"}, 0},
		{"help flag", []string{"-h"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			captureStdout(t, func() { got = run(tt.args) })
			if got != tt.want {
				t.Errorf("run(%v) = %d, want %d", tt.args, got, tt.want)
			}
		})
	}
}

func TestRunVersionPrints(t *testing.T) {
	out := captureStdout(t, func() { run([]string{"version"}) })
	if !strings.Contains(out, "vale ") {
		t.Errorf("version output = %q", out)
	}
}

func TestRunDefaultsToLint(t *testing.T) {
	dir := t.TempDir()
	// A file with a contraction: an error-severity finding, so the default
	// action (lint) reaches the gate and returns exit code 1.
	f := filepath.Join(dir, "doc.md")
	mustWrite(t, f, "Do not use words like don't here.\n")
	var code int
	out := captureStdout(t, func() { code = run([]string{f}) })
	if code != 1 {
		t.Fatalf("run(%q) = %d, want 1 (default lint, gate on error)", f, code)
	}
	// The finding appears in the concise report with an actionable location.
	if !strings.Contains(out, f+":1:") || !strings.Contains(out, "STE.Contractions") {
		t.Errorf("default-lint output not actionable: %q", out)
	}
}

func TestCmdLintAuditAlwaysExitsZero(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "Don't stop.") // a contraction normally exits 1
	var code int
	out := captureStdout(t, func() { code = cmdLint([]string{"-audit", f}) })
	if code != 0 {
		t.Errorf("audit exit = %d, want 0", code)
	}
	if !strings.Contains(out, "STE.Contractions") {
		t.Errorf("audit should still report findings: %q", out)
	}
}

func TestRunCleanFileDefaultsToLint(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "clean.md")
	mustWrite(t, f, "Attach the part.\n")
	var code int
	captureStdout(t, func() { code = run([]string{f}) })
	if code != 0 {
		t.Errorf("clean file via default lint = %d, want 0", code)
	}
}

func TestCmdRules(t *testing.T) {
	var code int
	out := captureStdout(t, func() { code = run([]string{"rules"}) })
	if code != 0 {
		t.Fatalf("rules exit = %d", code)
	}
	for _, want := range []string{"STE.Vocabulary", "STE.Contractions", "SEVERITY"} {
		if !strings.Contains(out, want) {
			t.Errorf("rules output missing %q\n%s", want, out)
		}
	}
}

func TestCmdLintNoPaths(t *testing.T) {
	if got := cmdLint(nil); got != 2 {
		t.Errorf("cmdLint(nil) = %d, want 2", got)
	}
}

func TestCmdLintBadFlag(t *testing.T) {
	if got := cmdLint([]string{"-nope"}); got != 2 {
		t.Errorf("bad flag = %d, want 2", got)
	}
}

func TestCmdLintBadFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "hi")
	if got := cmdLint([]string{"-format", "xml", f}); got != 2 {
		t.Errorf("bad format = %d, want 2", got)
	}
}

func TestCmdLintBadMarkdown(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "hi")
	if got := cmdLint([]string{"-markdown", "maybe", f}); got != 2 {
		t.Errorf("bad markdown = %d, want 2", got)
	}
}

func TestCmdLintMissingFile(t *testing.T) {
	if got := cmdLint([]string{filepath.Join(t.TempDir(), "nope.txt")}); got != 2 {
		t.Errorf("missing file = %d, want 2", got)
	}
}

func TestCmdLintBadConfig(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "hi")
	badCfg := filepath.Join(dir, "cfg.yml")
	mustWrite(t, badCfg, "minSeverity: bogus\n")
	if got := cmdLint([]string{"-config", badCfg, f}); got != 2 {
		t.Errorf("bad config = %d, want 2", got)
	}
}

func TestCmdLintCleanFileExitsZero(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "Open the door.")
	var code int
	out := captureStdout(t, func() { code = cmdLint([]string{f}) })
	if code != 0 {
		t.Errorf("clean file exit = %d, want 0\n%s", code, out)
	}
	// A clean file writes no findings to stdout; the summary goes to stderr.
	if out != "" {
		t.Errorf("clean file stdout should be empty, got %q", out)
	}
}

func TestCmdLintErrorFileExitsOne(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "Don't stop.")
	var code int
	out := captureStdout(t, func() { code = cmdLint([]string{f}) })
	if code != 1 {
		t.Errorf("error file exit = %d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "STE.Contractions") {
		t.Errorf("expected contraction in output: %q", out)
	}
}

func TestCmdLintJSONFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "Don't stop.")
	var code int
	out := captureStdout(t, func() { code = cmdLint([]string{"-format", "json", f}) })
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, `"results"`) || !strings.Contains(out, `"ruleId"`) {
		t.Errorf("json output = %q", out)
	}
}

func TestCmdLintStrictAndMinSeverity(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	// "bolt" is a strict-only bare word (suggestion). With min-severity
	// suggestion the gate fails (exit 1).
	mustWrite(t, f, "Open the bolt here.")
	var code int
	captureStdout(t, func() {
		code = cmdLint([]string{"-strict-vocabulary", "-min-severity", "suggestion", f})
	})
	if code != 1 {
		t.Errorf("strict suggestion gate exit = %d, want 1", code)
	}
}

func TestParseMarkdownMode(t *testing.T) {
	tests := []struct {
		in      string
		want    linter.MarkdownMode
		wantErr bool
	}{
		{"auto", linter.MarkdownAuto, false},
		{"", linter.MarkdownAuto, false},
		{"on", linter.MarkdownOn, false},
		{"true", linter.MarkdownOn, false},
		{"off", linter.MarkdownOff, false},
		{"false", linter.MarkdownOff, false},
		{"ON", linter.MarkdownOn, false},
		{"bogus", linter.MarkdownAuto, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseMarkdownMode(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("mode = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectFiles(t *testing.T) {
	dir := t.TempDir()
	// Files with known and unknown extensions.
	mustWrite(t, filepath.Join(dir, "b.md"), "x")
	mustWrite(t, filepath.Join(dir, "a.txt"), "x")
	mustWrite(t, filepath.Join(dir, "skip.go"), "x")
	sub := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(sub, "c.md"), "x") // should be skipped

	files, err := collectFiles([]string{dir}, pathFilter{})
	if err != nil {
		t.Fatal(err)
	}
	// Sorted, extension-filtered, node_modules skipped.
	want := []string{filepath.Join(dir, "a.txt"), filepath.Join(dir, "b.md")}
	if len(files) != len(want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("file %d = %q, want %q", i, files[i], want[i])
		}
	}
}

func TestCollectFilesDedupesExplicit(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWrite(t, f, "x")
	files, err := collectFiles([]string{f, f}, pathFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("files = %v, want one (deduped)", files)
	}
}

func TestCollectFilesMissing(t *testing.T) {
	if _, err := collectFiles([]string{filepath.Join(t.TempDir(), "nope")}, pathFilter{}); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestGateFailed(t *testing.T) {
	results := []report.FileResult{
		{Path: "a", Findings: []lint.Finding{{Severity: lint.SeverityWarning}}},
	}
	if !gateFailed(results, lint.SeverityWarning) {
		t.Error("warning should fail a warning gate")
	}
	if gateFailed(results, lint.SeverityError) {
		t.Error("warning should not fail an error gate")
	}
	if gateFailed(nil, lint.SeveritySuggestion) {
		t.Error("no findings should not fail")
	}
}

func TestCmdGen(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "wordset.json")
	out := filepath.Join(dir, "out.go")
	mustWrite(t, in, `{"set_name":"s","words":[],"alternatives":[{"title":"able","alt_title":"can"}]}`)
	var code int
	stdout := captureStdout(t, func() { code = cmdGen([]string{"-in", in, "-out", out}) })
	if code != 0 {
		t.Fatalf("cmdGen exit = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "wrote 1 substitutions") {
		t.Errorf("stdout = %q", stdout)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("output not written: %v", err)
	}
}

func TestCmdGenBadInput(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.go")
	if got := cmdGen([]string{"-in", filepath.Join(dir, "nope.json"), "-out", out}); got != 1 {
		t.Errorf("cmdGen bad input = %d, want 1", got)
	}
}

func TestCmdGenBadFlag(t *testing.T) {
	if got := cmdGen([]string{"-bogus"}); got != 2 {
		t.Errorf("cmdGen bad flag = %d, want 2", got)
	}
}

func TestCmdMCPBadConfig(t *testing.T) {
	dir := t.TempDir()
	badCfg := filepath.Join(dir, "cfg.yml")
	mustWrite(t, badCfg, "minSeverity: bogus\n")
	if got := cmdMCP([]string{"-config", badCfg}); got != 2 {
		t.Errorf("cmdMCP bad config = %d, want 2", got)
	}
}

func TestCmdMCPBadFlag(t *testing.T) {
	if got := cmdMCP([]string{"-bogus"}); got != 2 {
		t.Errorf("cmdMCP bad flag = %d, want 2", got)
	}
}

func TestCmdMCPServesUntilEOF(t *testing.T) {
	// Replace stdin with a closed/empty pipe so Serve reads EOF immediately and
	// returns cleanly with exit code 0.
	oldIn, oldOut := os.Stdin, os.Stdout
	rIn, wIn, _ := os.Pipe()
	wIn.Close() // immediate EOF
	_, wOut, _ := os.Pipe()
	os.Stdin = rIn
	os.Stdout = wOut
	defer func() { os.Stdin, os.Stdout = oldIn, oldOut; wOut.Close() }()

	code := cmdMCP(nil)
	if code != 0 {
		t.Errorf("cmdMCP EOF exit = %d, want 0", code)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPathFilterExcludeAndInclude(t *testing.T) {
	f := pathFilter{exclude: []string{"vendor/*", "*.gen.md"}, include: []string{"docs/*.md"}}
	cases := map[string]bool{
		"docs/guide.md":   true,  // included
		"vendor/x.md":     false, // excluded dir glob
		"docs/api.gen.md": false, // excluded by *.gen.md (base match)
		"other/x.md":      false, // not in include set
	}
	for path, want := range cases {
		if got := f.allows(path); got != want {
			t.Errorf("allows(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestPathFilterDoublestarBase(t *testing.T) {
	f := pathFilter{exclude: []string{"**/CHANGELOG.md"}}
	if !f.excludes("a/b/c/CHANGELOG.md") {
		t.Errorf("**/ pattern should match nested base name")
	}
	if f.excludes("a/b/README.md") {
		t.Errorf("should not exclude non-matching file")
	}
}

func TestReadValeIgnore(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".valeignore"), "# comment\n\nvendor/*\n*.tmp.md\n")
	pats := readValeIgnore(dir)
	if len(pats) != 2 || pats[0] != "vendor/*" || pats[1] != "*.tmp.md" {
		t.Errorf("readValeIgnore = %v, want [vendor/* *.tmp.md]", pats)
	}
}
