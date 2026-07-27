package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/linter"
)

// tool describes one MCP tool for the tools/list response.
type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// toolsList returns the tools that the server offers.
func (s *Server) toolsList() any {
	return map[string]any{
		"tools": []tool{
			{
				Name:        "lint_text",
				Description: "Check text against the Simplified Technical English rules and return the findings.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "The text to check.",
						},
						"filename": map[string]any{
							"type":        "string",
							"description": "An optional file name. A name that ends in .md turns on Markdown mode.",
						},
						"markdown": map[string]any{
							"type":        "boolean",
							"description": "Set Markdown mode directly. This overrides the file name.",
						},
						"minSeverity": map[string]any{
							"type":        "string",
							"enum":        []string{"error", "warning", "suggestion"},
							"description": "Return only findings at this severity or higher.",
						},
					},
					"required": []string{"text"},
				},
			},
			{
				Name:        "list_rules",
				Description: "List the Simplified Technical English rules that the linter uses.",
				InputSchema: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}
}

// callParams is the shape of a tools/call request.
type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// toolsCall runs a tool and returns the result content.
func (s *Server) toolsCall(raw json.RawMessage) (any, *rpcError) {
	var p callParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "invalid tool call: " + err.Error()}
	}
	switch p.Name {
	case "lint_text":
		return s.callLintText(p.Arguments)
	case "list_rules":
		return s.callListRules()
	default:
		return nil, &rpcError{Code: codeInvalidParams, Message: "unknown tool: " + p.Name}
	}
}

// lintTextArgs is the shape of the lint_text arguments.
type lintTextArgs struct {
	Text        string `json:"text"`
	Filename    string `json:"filename"`
	Markdown    *bool  `json:"markdown"`
	MinSeverity string `json:"minSeverity"`
}

// callLintText runs the lint_text tool.
func (s *Server) callLintText(raw json.RawMessage) (any, *rpcError) {
	var args lintTextArgs
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: "invalid arguments: " + err.Error()}
		}
	}
	if args.Text == "" {
		return toolError("The text argument is empty."), nil
	}
	mode := linter.MarkdownAuto
	if args.Markdown != nil {
		if *args.Markdown {
			mode = linter.MarkdownOn
		} else {
			mode = linter.MarkdownOff
		}
	}
	filename := args.Filename
	if filename == "" {
		filename = "text"
	}
	findings := s.linter.LintText(filename, args.Text, mode)
	if args.MinSeverity != "" {
		min, err := lint.ParseSeverity(args.MinSeverity)
		if err != nil {
			return toolError(err.Error()), nil
		}
		findings = filterSeverity(findings, min)
	}
	return toolResultJSON(map[string]any{
		"findings": findings,
		"summary":  summarize(findings),
	}, summaryLine(findings))
}

// callListRules runs the list_rules tool.
func (s *Server) callListRules() (any, *rpcError) {
	type ruleInfo struct {
		ID              string `json:"id"`
		Description     string `json:"description"`
		DefaultSeverity string `json:"defaultSeverity"`
	}
	var infos []ruleInfo
	for _, r := range s.linter.Rules() {
		infos = append(infos, ruleInfo{
			ID:              r.ID(),
			Description:     r.Description(),
			DefaultSeverity: string(r.DefaultSeverity()),
		})
	}
	return toolResultJSON(map[string]any{"rules": infos}, fmt.Sprintf("%d rules", len(infos)))
}

// filterSeverity keeps findings at or above the minimum severity.
func filterSeverity(findings []lint.Finding, min lint.Severity) []lint.Finding {
	out := make([]lint.Finding, 0, len(findings))
	for _, f := range findings {
		if f.Severity.AtLeast(min) {
			out = append(out, f)
		}
	}
	return out
}

// summarize counts findings by severity.
func summarize(findings []lint.Finding) map[string]int {
	m := map[string]int{"error": 0, "warning": 0, "suggestion": 0}
	for _, f := range findings {
		m[string(f.Severity)]++
	}
	return m
}

// summaryLine builds a short human summary of the findings.
func summaryLine(findings []lint.Finding) string {
	m := summarize(findings)
	return fmt.Sprintf("%d findings (%d errors, %d warnings, %d suggestions)",
		len(findings), m["error"], m["warning"], m["suggestion"])
}

// toolResultJSON builds a tool result with a text summary and a JSON block.
func toolResultJSON(payload any, summary string) (any, *rpcError) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, &rpcError{Code: codeInternalError, Message: err.Error()}
	}
	text := summary + "\n" + string(data)
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
		"isError": false,
	}, nil
}

// toolError builds a tool result that reports a user error.
func toolError(message string) any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
		"isError": true,
	}
}
