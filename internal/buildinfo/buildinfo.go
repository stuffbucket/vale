// Package buildinfo exposes the binary's version provenance: the semantic
// version, the source branch and commit, and the build timestamp. A release or
// Makefile build injects these with -ldflags; a plain "go build" or "go install"
// falls back to the VCS data that the Go toolchain embeds in the binary.
package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// These are set at build time with -ldflags "-X <pkg>.<name>=<value>". They stay
// unexported: callers read the resolved values through Get.
var (
	version = "dev"
	commit  = ""
	branch  = ""
	date    = ""
)

// Info is the resolved build provenance.
type Info struct {
	Version string // semantic version, without a leading "v" (for example "0.2.0")
	Branch  string // source branch, or "unknown"
	Commit  string // short commit hash, or "unknown"
	Date    string // build timestamp (RFC 3339, approximate), or "unknown"
	Dirty   bool   // the work tree had uncommitted changes at build time
}

// Get resolves the build provenance. It fills each gap that the linker left from
// the VCS metadata that the Go toolchain embeds ("go build" or "go install").
func Get() Info {
	bi, _ := debug.ReadBuildInfo()
	return resolve(Info{
		Version: normalizeVersion(version),
		Branch:  branch,
		Commit:  shortHash(commit),
		Date:    date,
	}, bi)
}

// resolve merges the linker values with the embedded VCS build info. It is split
// from Get so that a test can drive the fallback logic without a real build.
func resolve(in Info, bi *debug.BuildInfo) Info {
	if bi != nil {
		if isUnset(in.Version) && isRealModuleVersion(bi.Main.Version) {
			in.Version = normalizeVersion(bi.Main.Version)
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if in.Commit == "" {
					in.Commit = shortHash(s.Value)
				}
			case "vcs.time":
				if in.Date == "" {
					in.Date = s.Value
				}
			case "vcs.modified":
				if s.Value == "true" {
					in.Dirty = true
				}
			}
		}
	}
	if in.Version == "" {
		in.Version = "dev"
	}
	in.Branch = orUnknown(in.Branch)
	in.Commit = orUnknown(in.Commit)
	in.Date = orUnknown(in.Date)
	return in
}

func isUnset(v string) bool { return v == "" || v == "dev" }

func isRealModuleVersion(v string) bool { return v != "" && v != "(devel)" }

// normalizeVersion drops a leading "v" and any trailing dirty marker (git
// describe adds "-dirty"; the Go toolchain adds "+dirty"). The Dirty flag, read
// from vcs.modified, is the single source for the dirty state in the banner.
func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSuffix(v, "-dirty")
	v = strings.TrimSuffix(v, "+dirty")
	return v
}

func orUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}

func shortHash(h string) string {
	const n = 12
	if len(h) > n {
		return h[:n]
	}
	return h
}

// String is the one-line version banner, for example:
// "vale 0.2.0 (branch main, commit a1b2c3d4e5f6, built 2026-07-29T10:16:21Z)".
func (i Info) String() string {
	s := fmt.Sprintf("vale %s (branch %s, commit %s, built %s)", i.Version, i.Branch, i.Commit, i.Date)
	if i.Dirty {
		s += " (dirty)"
	}
	return s
}
