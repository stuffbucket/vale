package buildinfo

import (
	"runtime/debug"
	"testing"
)

func TestResolveUsesLinkerValuesFirst(t *testing.T) {
	t.Parallel()
	in := Info{Version: "0.2.0", Branch: "main", Commit: "a1b2c3d4e5f6", Date: "2026-07-29T10:16:21Z"}
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v9.9.9"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "ffffffffffffffff"},
			{Key: "vcs.time", Value: "2000-01-01T00:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	got := resolve(in, bi)
	// Linker-injected values win; only Dirty comes from the embedded VCS info.
	if got.Version != "0.2.0" || got.Branch != "main" || got.Commit != "a1b2c3d4e5f6" || got.Date != "2026-07-29T10:16:21Z" {
		t.Fatalf("linker values overridden: %+v", got)
	}
	if !got.Dirty {
		t.Fatalf("expected Dirty from vcs.modified, got %+v", got)
	}
}

func TestResolveFallsBackToVCS(t *testing.T) {
	t.Parallel()
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123"},
			{Key: "vcs.time", Value: "2026-07-29T10:00:00Z"},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	got := resolve(Info{Version: "dev"}, bi)
	if got.Commit != "0123456789ab" { // truncated to 12
		t.Errorf("commit = %q, want 0123456789ab", got.Commit)
	}
	if got.Date != "2026-07-29T10:00:00Z" {
		t.Errorf("date = %q", got.Date)
	}
	if got.Branch != "unknown" { // VCS info carries no branch
		t.Errorf("branch = %q, want unknown", got.Branch)
	}
	if got.Version != "dev" { // (devel) is not a real module version
		t.Errorf("version = %q, want dev", got.Version)
	}
	if got.Dirty {
		t.Errorf("Dirty should be false")
	}
}

func TestResolveModuleVersionFallback(t *testing.T) {
	t.Parallel()
	bi := &debug.BuildInfo{Main: debug.Module{Version: "v0.3.1"}}
	got := resolve(Info{Version: "dev"}, bi)
	if got.Version != "0.3.1" { // "go install module@v0.3.1", leading v trimmed
		t.Errorf("version = %q, want 0.3.1", got.Version)
	}
}

func TestResolveNilBuildInfoFillsUnknowns(t *testing.T) {
	t.Parallel()
	got := resolve(Info{Version: "dev"}, nil)
	if got.Branch != "unknown" || got.Commit != "unknown" || got.Date != "unknown" {
		t.Fatalf("gaps not filled: %+v", got)
	}
}

func TestStringFormat(t *testing.T) {
	t.Parallel()
	clean := Info{Version: "0.2.0", Branch: "main", Commit: "abc123", Date: "2026-07-29T10:16:21Z"}
	want := "vale 0.2.0 (branch main, commit abc123, built 2026-07-29T10:16:21Z)"
	if got := clean.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	dirty := clean
	dirty.Dirty = true
	if got := dirty.String(); got != want+" (dirty)" {
		t.Errorf("dirty String() = %q", got)
	}
}

func TestShortHash(t *testing.T) {
	t.Parallel()
	if got := shortHash("0123456789abcdef"); got != "0123456789ab" {
		t.Errorf("shortHash long = %q", got)
	}
	if got := shortHash("abc"); got != "abc" { // shorter than the limit is kept whole
		t.Errorf("shortHash short = %q", got)
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()
	if got := normalizeVersion("v1.2.3"); got != "1.2.3" {
		t.Errorf("normalizeVersion = %q", got)
	}
	if got := normalizeVersion("1.2.3"); got != "1.2.3" {
		t.Errorf("normalizeVersion no-v = %q", got)
	}
	if got := normalizeVersion("v0.1.0-dirty"); got != "0.1.0" { // git describe dirty marker
		t.Errorf("normalizeVersion describe-dirty = %q", got)
	}
	if got := normalizeVersion("v0.1.0+dirty"); got != "0.1.0" { // Go toolchain dirty marker
		t.Errorf("normalizeVersion toolchain-dirty = %q", got)
	}
	if got := normalizeVersion("v0.1.0-3-gabc1234"); got != "0.1.0-3-gabc1234" { // commits past a tag are kept
		t.Errorf("normalizeVersion post-tag = %q", got)
	}
}
