package eval

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/linter"
)

// slopPrefix marks the opt-in slop rule family.
const slopPrefix = "STE.Slop"

// Options configures a run.
type Options struct {
	Client      *Client
	Linter      *linter.Linter // must have the slop family enabled
	Models      []string       // explicit; empty means discover all from the endpoint
	Prompts     []string
	MaxTokens   int
	Concurrency int
}

// Sample is the outcome of one (model, prompt) request.
type Sample struct {
	Model, Family, Prompt string
	Output                string
	Words                 int
	Findings              []lint.Finding
	Err                   error
}

// ModelResult aggregates a model's samples.
type ModelResult struct {
	Model, Family                    string
	Samples, Errors, Words           int
	SlopFindings, TotalFindings      int
	SlopPer100Words, TotalPer100Word float64
	ByRule                           map[string]int
	FirstError                       string // first error message, for diagnosis
}

// FamilyResult aggregates a family's models.
type FamilyResult struct {
	Family                           string
	Models, Words                    int
	SlopFindings, TotalFindings      int
	SlopPer100Words, TotalPer100Word float64
}

// Report is the full result of a run.
type Report struct {
	Endpoint string
	Models   []ModelResult
	Families []FamilyResult
	Samples  []Sample
}

// Run drives every (model, prompt) pair concurrently, lints each reply, and
// aggregates the slop metrics. Model families come from the endpoint's
// /v1/models `owned_by` field.
func Run(ctx context.Context, opts Options) (*Report, error) {
	familyOf := map[string]string{}
	models := opts.Models
	discovered, err := opts.Client.Models(ctx)
	if err != nil && len(models) == 0 {
		return nil, fmt.Errorf("discover models: %w", err)
	}
	for _, m := range discovered {
		familyOf[m.ID] = m.OwnedBy
	}
	if len(models) == 0 {
		for _, m := range discovered {
			models = append(models, m.ID)
		}
	}
	prompts := opts.Prompts
	if len(prompts) == 0 {
		prompts = DefaultPrompts
	}

	type task struct{ model, prompt string }
	var tasks []task
	for _, m := range models {
		for _, p := range prompts {
			tasks = append(tasks, task{m, p})
		}
	}

	samples := make([]Sample, len(tasks))
	sem := make(chan struct{}, max(1, opts.Concurrency))
	var wg sync.WaitGroup
	for i, t := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, t task) {
			defer wg.Done()
			defer func() { <-sem }()
			s := Sample{Model: t.model, Family: familyForID(familyOf, t.model), Prompt: t.prompt}
			out, err := opts.Client.Complete(ctx, t.model, t.prompt, opts.MaxTokens)
			if err != nil {
				s.Err = err
			} else {
				s.Output = out
				s.Words = len(strings.Fields(out))
				s.Findings = opts.Linter.LintText("output.md", out, linter.MarkdownOn)
			}
			samples[i] = s
		}(i, t)
	}
	wg.Wait()

	return aggregate(opts.Client.BaseURL, samples), nil
}

// familyForID returns the family, defaulting to "unknown".
func familyForID(m map[string]string, id string) string {
	if f, ok := m[id]; ok && f != "" {
		return f
	}
	return "unknown"
}

// aggregate rolls samples up into per-model and per-family results.
func aggregate(endpoint string, samples []Sample) *Report {
	rep := &Report{Endpoint: endpoint, Samples: samples}
	byModel := map[string]*ModelResult{}
	var modelOrder []string
	for _, s := range samples {
		mr, ok := byModel[s.Model]
		if !ok {
			mr = &ModelResult{Model: s.Model, Family: s.Family, ByRule: map[string]int{}}
			byModel[s.Model] = mr
			modelOrder = append(modelOrder, s.Model)
		}
		mr.Samples++
		if s.Err != nil {
			mr.Errors++
			if mr.FirstError == "" {
				mr.FirstError = truncate(s.Err.Error(), 160)
			}
			continue
		}
		mr.Words += s.Words
		for _, f := range s.Findings {
			mr.TotalFindings++
			mr.ByRule[f.RuleID]++
			if strings.HasPrefix(f.RuleID, slopPrefix) {
				mr.SlopFindings++
			}
		}
	}
	byFamily := map[string]*FamilyResult{}
	var familyOrder []string
	for _, id := range modelOrder {
		mr := byModel[id]
		mr.SlopPer100Words = per100(mr.SlopFindings, mr.Words)
		mr.TotalPer100Word = per100(mr.TotalFindings, mr.Words)
		rep.Models = append(rep.Models, *mr)

		fr, ok := byFamily[mr.Family]
		if !ok {
			fr = &FamilyResult{Family: mr.Family}
			byFamily[mr.Family] = fr
			familyOrder = append(familyOrder, mr.Family)
		}
		fr.Models++
		fr.Words += mr.Words
		fr.SlopFindings += mr.SlopFindings
		fr.TotalFindings += mr.TotalFindings
	}
	for _, fam := range familyOrder {
		fr := byFamily[fam]
		fr.SlopPer100Words = per100(fr.SlopFindings, fr.Words)
		fr.TotalPer100Word = per100(fr.TotalFindings, fr.Words)
		rep.Families = append(rep.Families, *fr)
	}
	// Rank families by slop density, most slop first — the headline result.
	sort.SliceStable(rep.Families, func(i, j int) bool {
		return rep.Families[i].SlopPer100Words > rep.Families[j].SlopPer100Words
	})
	return rep
}

// per100 is findings per 100 words, guarding against a zero denominator.
func per100(count, words int) float64 {
	if words == 0 {
		return 0
	}
	return float64(count) * 100 / float64(words)
}

// truncate shortens a string for display.
func truncate(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
