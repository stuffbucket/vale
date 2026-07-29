package eval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
)

// mockEndpoint serves /v1/models and /v1/chat/completions. The completion echoes
// a canned reply chosen by the model id so a test can assert slop scoring.
func mockEndpoint(t *testing.T, replies map[string]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []Model{
			{ID: "sloppy-1", OwnedBy: "OpenAI"},
			{ID: "plain-1", OwnedBy: "Anthropic"},
		}})
	})
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		reply := replies[body.Model]
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": reply}}},
		})
	})
	return httptest.NewServer(mux)
}

func TestClientModelsAndComplete(t *testing.T) {
	srv := mockEndpoint(t, map[string]string{"sloppy-1": "delve into the intricate realm"})
	defer srv.Close()
	c := NewClient(srv.URL, "")
	models, err := c.Models(context.Background())
	if err != nil || len(models) != 2 {
		t.Fatalf("models = %+v, err %v", models, err)
	}
	if models[0].OwnedBy != "OpenAI" {
		t.Errorf("family = %q", models[0].OwnedBy)
	}
	out, err := c.Complete(context.Background(), "sloppy-1", "x", 0, nil)
	if err != nil || !strings.Contains(out, "delve") {
		t.Fatalf("complete = %q, err %v", out, err)
	}
}

func TestRunScoresSlopByFamily(t *testing.T) {
	srv := mockEndpoint(t, map[string]string{
		// Sloppy output: several watchlist words.
		"sloppy-1": "We delve into the intricate realm to showcase a groundbreaking, meticulous design.",
		// Plain output: no slop words.
		"plain-1": "Open the valve. Attach the bracket. Close the cover.",
	})
	defer srv.Close()

	cfg := config.Default()
	cfg.Slop.Enabled = true
	rep, err := Run(context.Background(), Options{
		Client:      NewClient(srv.URL, ""),
		Linter:      linter.New(cfg),
		Prompts:     []string{"one"},
		Concurrency: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Models) != 2 || len(rep.Families) != 2 {
		t.Fatalf("expected 2 models and 2 families: %+v", rep)
	}
	// The sloppy model's family must rank first (most slop per 100 words).
	if rep.Families[0].Family != "OpenAI" {
		t.Errorf("expected OpenAI (sloppy) ranked first, got %+v", rep.Families)
	}
	if rep.Families[0].SlopPer100Words <= rep.Families[1].SlopPer100Words {
		t.Errorf("sloppy family should out-score plain: %+v", rep.Families)
	}
	// The plain model should score zero slop.
	for _, m := range rep.Models {
		if m.Model == "plain-1" && m.SlopFindings != 0 {
			t.Errorf("plain model scored slop: %+v", m)
		}
	}
}
