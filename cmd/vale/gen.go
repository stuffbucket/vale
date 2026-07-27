package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stuffbucket/vale/internal/vocab/generator"
)

// Default paths for the generator, relative to the module root.
const (
	defaultWordset = "third_party/openste/openste.json"
	defaultOutput  = "internal/vocab/substitutions_gen.go"
)

// cmdGen runs the gen subcommand. It builds the vocabulary source from the
// wordset. This is the same work that "go generate" does.
func cmdGen(args []string) int {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	in := fs.String("in", "", "path to the wordset JSON (default: the vendored wordset)")
	out := fs.String("out", "", "path to the generated source (default: the vocab package)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	inPath := *in
	outPath := *out
	if inPath == "" || outPath == "" {
		root, err := moduleRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gen: %v\n", err)
			return 2
		}
		if inPath == "" {
			inPath = filepath.Join(root, defaultWordset)
		}
		if outPath == "" {
			outPath = filepath.Join(root, defaultOutput)
		}
	}

	n, err := generator.Generate(inPath, outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen: %v\n", err)
		return 1
	}
	fmt.Printf("wrote %d substitutions to %s\n", n, outPath)
	return 0
}

// moduleRoot walks up from the working directory and returns the first
// directory that holds a go.mod file.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found; give --in and --out")
		}
		dir = parent
	}
}
