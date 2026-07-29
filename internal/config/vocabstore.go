package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// DefaultVocabStore is the project-scoped vocab store file name, discovered by
// config layering. The MCP defaults to the user-level store (see
// DefaultVocabStorePath); a project can pin a local one with this name.
const DefaultVocabStore = ".vale-ste.vocab.yml"

// vocabStoreNames are the discovered project vocab-store file names. They sit
// just below an explicit --config, so learned terms outrank project config.
var vocabStoreNames = []string{".vale-ste.vocab.yml", ".vale-ste.vocab.yaml"}

// xdgStateHome returns $XDG_STATE_HOME, or ~/.local/state, or "".
func xdgStateHome() string {
	if v := os.Getenv("XDG_STATE_HOME"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "state")
}

// StateDir is vale's XDG state directory (where learned data lives).
func StateDir() string {
	base := xdgStateHome()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "vale-ste")
}

// DefaultVocabStorePath is where the MCP persists learned vocabulary by default:
// a user-level XDG state file that survives restarts and is shared across
// sessions and projects. It falls back to the project-scoped name when there is
// no home directory.
func DefaultVocabStorePath() string {
	if dir := StateDir(); dir != "" {
		return filepath.Join(dir, "vocab.yml")
	}
	return DefaultVocabStore
}

// vocabStoreDoc is the on-disk shape of a vocab store.
type vocabStoreDoc struct {
	Vocabulary struct {
		Allow []string `yaml:"allow"`
		Deny  []string `yaml:"deny"`
	} `yaml:"vocabulary"`
}

// ReadVocabStore returns the allow and deny terms in a store file. A missing
// file yields empty slices and no error.
func ReadVocabStore(path string) (allow, deny []string, err error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read vocab store %s: %w", path, err)
	}
	var doc vocabStoreDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse vocab store %s: %w", path, err)
	}
	return doc.Vocabulary.Allow, doc.Vocabulary.Deny, nil
}

// UpdateVocabStore merges new allow and deny terms into the store file (creating
// it if needed) and returns the merged, sorted sets. It rewrites the file
// wholesale — vale owns this file.
func UpdateVocabStore(path string, addAllow, addDeny []string) (allow, deny []string, err error) {
	curAllow, curDeny, err := ReadVocabStore(path)
	if err != nil {
		return nil, nil, err
	}
	allow = mergeTerms(curAllow, addAllow)
	deny = mergeTerms(curDeny, addDeny)

	var doc vocabStoreDoc
	doc.Vocabulary.Allow = allow
	doc.Vocabulary.Deny = deny

	var buf bytes.Buffer
	buf.WriteString("# Managed by vale — terms learned through the MCP update_vocabulary tool.\n")
	buf.WriteString("# Edit vale-ste config files by hand; this one is rewritten.\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, nil, err
	}
	_ = enc.Close()
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, fmt.Errorf("create vocab store dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return nil, nil, fmt.Errorf("write vocab store %s: %w", path, err)
	}
	return allow, deny, nil
}

// mergeTerms unions two term lists, trims and lower-cases via normalizeTerm,
// drops empties, and returns a sorted, de-duplicated slice.
func mergeTerms(cur, add []string) []string {
	seen := map[string]bool{}
	for _, t := range append(append([]string{}, cur...), add...) {
		if n := normalizeTerm(t); n != "" {
			seen[n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
