package fix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/eval"
	"github.com/stuffbucket/vale/internal/lint"
)

func TestFixSendsFindingsAndReturnsRewrite(t *testing.T) {
	var gotPrompt string
	var gotTemp *float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages    []eval.Message `json:"messages"`
			Temperature *float64       `json:"temperature"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.Messages) > 0 {
			gotPrompt = body.Messages[0].Content
		}
		gotTemp = body.Temperature
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "```markdown\nThe unit is under test.\n```"}}},
		})
	}))
	defer srv.Close()

	temp := 0.3
	out, err := Fix(context.Background(), Options{
		Client: eval.NewClient(srv.URL, ""),
		Model:  "m",
		Path:   "doc.md",
		Text:   "The unit is being tested.",
		Findings: []lint.Finding{
			{RuleID: "STE.PassiveVoice", Severity: lint.SeverityWarning, Message: "passive", Line: 1, Col: 13, Match: "being tested"},
		},
		Temperature: &temp,
	})
	if err != nil {
		t.Fatal(err)
	}
	// The wrapping code fence is stripped.
	if out != "The unit is under test." {
		t.Errorf("out = %q", out)
	}
	// The prompt carried the finding and the document.
	if !strings.Contains(gotPrompt, "STE.PassiveVoice") || !strings.Contains(gotPrompt, "being tested") {
		t.Errorf("prompt missing findings: %q", gotPrompt)
	}
	// The temperature override reached the request.
	if gotTemp == nil || *gotTemp != 0.3 {
		t.Errorf("temperature = %v, want 0.3", gotTemp)
	}
}

func TestFixRequiresModel(t *testing.T) {
	_, err := Fix(context.Background(), Options{Client: eval.NewClient("http://x", ""), Text: "hi"})
	if err == nil {
		t.Fatal("expected an error when no model is set")
	}
}

func TestStripFence(t *testing.T) {
	if got := stripFence("```md\nbody\n```"); got != "body" {
		t.Errorf("stripFence = %q", got)
	}
	if got := stripFence("no fence here"); got != "no fence here" {
		t.Errorf("stripFence should leave unfenced text: %q", got)
	}
}
