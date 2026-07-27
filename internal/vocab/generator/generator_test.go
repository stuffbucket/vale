package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleJSON = `{
  "set_name": "test set",
  "words": [
    {"title": "able", "wordstatus": "unapproved"},
    {"title": "foo", "wordstatus": "unapproved"},
    {"title": "bar", "wordstatus": "approved"},
    {"title": "baz (note)", "wordstatus": "unapproved"}
  ],
  "alternatives": [
    {"title": "able", "alt_title": "can"},
    {"title": "able", "alt_title": "possible"},
    {"title": "able", "alt_title": "can"},
    {"title": "apple", "alt_title": "seed"},
    {"title": "banana", "alt_title": "fruit"},
    {"title": "set up", "alt_title": "install"},
    {"title": "zebra", "alt_title": "zebra"},
    {"title": "bad (note)", "alt_title": "x"},
    {"title": "empty", "alt_title": ""}
  ]
}`

func TestBuildGroupingSortingFiltering(t *testing.T) {
	subs, err := Build([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	// zebra (self-referential), bad (note) (parenthetical), and empty (no alt)
	// are all dropped. Remaining are sorted by word.
	wantWords := []string{"able", "apple", "banana", "set up"}
	if len(subs) != len(wantWords) {
		t.Fatalf("subs = %d (%+v), want %d", len(subs), subs, len(wantWords))
	}
	for i, w := range wantWords {
		if subs[i].Word != w {
			t.Errorf("sub %d word = %q, want %q", i, subs[i].Word, w)
		}
	}
	// "able" dedups the repeated "can" and keeps insertion order.
	able := subs[0]
	if len(able.Alternatives) != 2 || able.Alternatives[0] != "can" || able.Alternatives[1] != "possible" {
		t.Errorf("able alternatives = %v, want [can possible]", able.Alternatives)
	}
}

func TestBuildInvalidJSON(t *testing.T) {
	if _, err := Build([]byte("{not json")); err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestRender(t *testing.T) {
	subs, err := Build([]byte(sampleJSON))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render("my set", subs, []string{"foo", "gamma"})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	src := string(out)
	for _, want := range []string{
		"package vocab",
		"DO NOT EDIT",
		"my set",
		"var Substitutions = []Substitution{",
		`{Word: "able", Alternatives: []string{"can", "possible"}}`,
		"var UnapprovedWords = []string{",
		`"foo"`,
		`"gamma"`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("rendered source missing %q\n---\n%s", want, src)
		}
	}
}

func TestRenderIsGofmtStable(t *testing.T) {
	// Render runs format.Source; the output should be idempotent when re-rendered.
	subs, _ := Build([]byte(sampleJSON))
	a, err := Render("x", subs, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Render("x", subs, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Error("Render is not deterministic")
	}
}

func TestGenerateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "wordset.json")
	out := filepath.Join(dir, "out.go")
	if err := os.WriteFile(in, []byte(sampleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Generate(in, out)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if n != 4 {
		t.Errorf("substitution count = %d, want 4", n)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	if !strings.Contains(src, `Word: "able"`) {
		t.Errorf("generated file missing able sub")
	}
	// "foo" is unapproved with no substitution; "able" has a sub so it is excluded
	// from the bare list.
	uIdx := strings.Index(src, "var UnapprovedWords")
	if uIdx < 0 {
		t.Fatal("no UnapprovedWords block")
	}
	tail := src[uIdx:]
	if !strings.Contains(tail, `"foo"`) {
		t.Errorf("UnapprovedWords should contain foo:\n%s", tail)
	}
	if strings.Contains(tail, `"able"`) {
		t.Errorf("UnapprovedWords should not contain able (it has a substitution):\n%s", tail)
	}
}

func TestGenerateMissingInput(t *testing.T) {
	if _, err := Generate(filepath.Join(t.TempDir(), "nope.json"), filepath.Join(t.TempDir(), "o.go")); err == nil {
		t.Fatal("expected error for missing input")
	}
}

func TestGenerateBadJSON(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(in, []byte("{oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(in, filepath.Join(dir, "o.go")); err == nil {
		t.Fatal("expected parse error")
	}
}

// TestCommittedFileIsUpToDate checks that the committed substitutions_gen.go is
// exactly what a fresh Generate of the vendored OpenSTE wordset produces.
func TestCommittedFileIsUpToDate(t *testing.T) {
	// The test runs in internal/vocab/generator; walk up to the repo paths.
	jsonPath := filepath.Join("..", "..", "..", "third_party", "openste", "openste.json")
	committedPath := filepath.Join("..", "substitutions_gen.go")

	if _, err := os.Stat(jsonPath); err != nil {
		t.Skipf("wordset not available: %v", err)
	}
	committed, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("read committed file: %v", err)
	}

	out := filepath.Join(t.TempDir(), "fresh.go")
	if _, err := Generate(jsonPath, out); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	fresh, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(fresh) != string(committed) {
		t.Errorf("committed substitutions_gen.go is stale; re-run 'vale gen' "+
			"(fresh %d bytes vs committed %d bytes)", len(fresh), len(committed))
	}
}
