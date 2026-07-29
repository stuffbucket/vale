// Package fix rewrites a document with an LLM so it resolves its own lint
// findings, instead of only reporting them. It sends the original text plus the
// concise findings to a model on the configured endpoint and returns the
// corrected document.
package fix

import (
	"context"
	"fmt"
	"strings"

	"github.com/stuffbucket/vale/internal/eval"
	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/report"
)

// Options configures a fix request.
type Options struct {
	Client      *eval.Client
	Model       string
	Path        string
	Text        string
	Findings    []lint.Finding
	MaxTokens   int
	Temperature *float64
}

// Fix asks the model to rewrite Text so it resolves Findings, and returns the
// corrected document (with any wrapping code fence stripped).
func Fix(ctx context.Context, o Options) (string, error) {
	if strings.TrimSpace(o.Model) == "" {
		return "", fmt.Errorf("no model set; use --model or the model.name config")
	}
	issues := report.ConciseString([]report.FileResult{{Path: o.Path, Findings: o.Findings}})
	if strings.TrimSpace(issues) == "" {
		issues = "(no findings; return the document unchanged apart from obvious plain-language improvements)"
	}
	out, err := o.Client.Complete(ctx, o.Model, buildPrompt(o.Text, issues), o.MaxTokens, o.Temperature)
	if err != nil {
		return "", err
	}
	return stripFence(strings.TrimSpace(out)), nil
}

// buildPrompt frames the rewrite task for the model.
func buildPrompt(text, issues string) string {
	var b strings.Builder
	b.WriteString("You are a technical editor. Rewrite the document below in Simplified ")
	b.WriteString("Technical English so it resolves the listed issues, while you keep the ")
	b.WriteString("original meaning, structure, headings, and Markdown. Use short sentences, ")
	b.WriteString("the active voice, no contractions, plain concrete vocabulary, and one ")
	b.WriteString("instruction per sentence. Do not add commentary or a preamble. Output only ")
	b.WriteString("the corrected document.\n\n=== ISSUES ===\n")
	b.WriteString(issues)
	b.WriteString("\n=== DOCUMENT ===\n")
	b.WriteString(text)
	return b.String()
}

// stripFence removes a single wrapping Markdown code fence, which some models
// add around a whole-document reply.
func stripFence(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		return s
	}
	if strings.TrimSpace(lines[len(lines)-1]) != "```" {
		return s
	}
	return strings.Join(lines[1:len(lines)-1], "\n")
}
