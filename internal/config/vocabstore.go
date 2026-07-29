package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// DefaultVocabStore is the file the MCP writes session-learned terms to. It is a
// dedicated, vale-managed config fragment. Because it is a discovered layer (see
// vocabStoreNames), the CLI and any hook honor learned terms too.
const DefaultVocabStore = ".vale-ste.vocab.yml"

// vocabStoreNames are the discovered vocab-store file names. They sit just below
// an explicit --config, so learned terms take precedence over project config.
var vocabStoreNames = []string{".vale-ste.vocab.yml", ".vale-ste.vocab.yaml"}

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
