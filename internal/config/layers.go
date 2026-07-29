// Config layering. vale resolves configuration the way Claude and other
// XDG-aware tools do: several files from broad to narrow scope, each one merged
// on top of the last. Lowest to highest precedence:
//
//  1. system   — $XDG_CONFIG_DIRS/vale-ste/config.yml (default /etc/xdg)
//  2. user     — $XDG_CONFIG_HOME/vale-ste/config.yml (default ~/.config)
//  3. project  — the nearest .vale-ste.yml walking up from the working directory
//  4. local    — the nearest .vale-ste.local.yml (personal, git-ignored)
//  5. explicit — a --config <file>, when given
//
// Scalars from a higher layer replace lower ones. vocabulary.allow and
// vocabulary.deny ACCUMULATE across layers, so a user-level allow list and a
// project-level allow list add together. Rule settings merge per rule id.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	userFileNames    = []string{"config.yml", "config.yaml"}
	projectFileNames = []string{".vale-ste.yml", ".vale-ste.yaml"}
	localFileNames   = []string{".vale-ste.local.yml", ".vale-ste.local.yaml"}
)

// fileConfig is one raw layer. Its fields are pointers (or slices) so the merge
// can tell "absent" from "set to the zero value" and only override on set.
type fileConfig struct {
	MinSeverity      *string                `yaml:"minSeverity"`
	StrictVocabulary *bool                  `yaml:"strictVocabulary"`
	Sentence         *sentenceFile          `yaml:"sentence"`
	Vocabulary       *vocabularyFile        `yaml:"vocabulary"`
	Rules            map[string]RuleSetting `yaml:"rules"`
}

type sentenceFile struct {
	ProcedureMax   *int `yaml:"procedureMax"`
	DescriptionMax *int `yaml:"descriptionMax"`
}

type vocabularyFile struct {
	Allow                 []string `yaml:"allow"`
	Deny                  []string `yaml:"deny"`
	BuiltinTechnicalTerms *bool    `yaml:"builtinTechnicalTerms"`
}

// discoverLayers returns the existing config files in precedence order, lowest
// first. dir is the starting directory for the project search; explicit is an
// optional --config path, always appended as the top layer.
func discoverLayers(dir, explicit string) []string {
	var layers []string
	// System dirs: earlier entries in XDG_CONFIG_DIRS outrank later ones, so
	// append in reverse (lowest precedence first).
	sysDirs := xdgConfigDirs()
	for i := len(sysDirs) - 1; i >= 0; i-- {
		if p := firstExisting(filepath.Join(sysDirs[i], "vale-ste"), userFileNames); p != "" {
			layers = append(layers, p)
		}
	}
	if home := xdgConfigHome(); home != "" {
		if p := firstExisting(filepath.Join(home, "vale-ste"), userFileNames); p != "" {
			layers = append(layers, p)
		}
	}
	if p := searchUp(dir, projectFileNames); p != "" {
		layers = append(layers, p)
	}
	if p := searchUp(dir, localFileNames); p != "" {
		layers = append(layers, p)
	}
	if explicit != "" {
		layers = append(layers, explicit)
	}
	return layers
}

// xdgConfigHome returns $XDG_CONFIG_HOME, or ~/.config, or "" when neither is
// available.
func xdgConfigHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

// xdgConfigDirs returns the $XDG_CONFIG_DIRS entries, defaulting to /etc/xdg.
func xdgConfigDirs() []string {
	v := os.Getenv("XDG_CONFIG_DIRS")
	if v == "" {
		v = "/etc/xdg"
	}
	var dirs []string
	for _, d := range filepath.SplitList(v) {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// firstExisting returns the first name in dir that is a file, or "".
func firstExisting(dir string, names []string) string {
	if dir == "" {
		return ""
	}
	for _, n := range names {
		p := filepath.Join(dir, n)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

// searchUp walks up from dir and returns the first matching file, or "".
func searchUp(dir string, names []string) string {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if p := firstExisting(abs, names); p != "" {
			return p
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}

// loadFileLayer reads and parses one layer file.
func loadFileLayer(path string) (*fileConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(raw, &fc); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &fc, nil
}

// mergeFile folds one layer into the accumulator. Scalars replace on set; the
// allow/deny lists accumulate; rule settings merge per id.
func mergeFile(acc *Config, fc *fileConfig) {
	if fc == nil {
		return
	}
	if fc.MinSeverity != nil {
		acc.MinSeverity = *fc.MinSeverity
	}
	if fc.StrictVocabulary != nil {
		acc.StrictVocabulary = *fc.StrictVocabulary
	}
	if fc.Sentence != nil {
		if fc.Sentence.ProcedureMax != nil {
			acc.Sentence.ProcedureMax = *fc.Sentence.ProcedureMax
		}
		if fc.Sentence.DescriptionMax != nil {
			acc.Sentence.DescriptionMax = *fc.Sentence.DescriptionMax
		}
	}
	if fc.Vocabulary != nil {
		acc.Vocabulary.Allow = append(acc.Vocabulary.Allow, fc.Vocabulary.Allow...)
		acc.Vocabulary.Deny = append(acc.Vocabulary.Deny, fc.Vocabulary.Deny...)
		if fc.Vocabulary.BuiltinTechnicalTerms != nil {
			acc.Vocabulary.BuiltinTechnicalTerms = fc.Vocabulary.BuiltinTechnicalTerms
		}
	}
	if len(fc.Rules) > 0 && acc.Rules == nil {
		acc.Rules = map[string]RuleSetting{}
	}
	for id, rs := range fc.Rules {
		acc.Rules[id] = rs
	}
}
