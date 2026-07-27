package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
)

// runServer feeds the input lines to a server and returns the decoded response
// objects (one per output line).
func runServer(t *testing.T, input string) []map[string]any {
	t.Helper()
	l := linter.New(config.Default())
	var out bytes.Buffer
	if err := Serve(context.Background(), strings.NewReader(input), &out, l, "1.2.3"); err != nil {
		t.Fatalf("Serve error: %v", err)
	}
	var msgs []map[string]any
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("decode %q: %v", line, err)
		}
		msgs = append(msgs, m)
	}
	return msgs
}

func TestServeFullSession(t *testing.T) {
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"lint_text","arguments":{"text":"Don't stop."}}}`,
	}, "\n") + "\n"

	msgs := runServer(t, input)
	// The notification produces no response, so there are exactly 3 responses.
	if len(msgs) != 3 {
		t.Fatalf("responses = %d, want 3: %v", len(msgs), msgs)
	}

	// initialize
	initResult := msgs[0]["result"].(map[string]any)
	serverInfo := initResult["serverInfo"].(map[string]any)
	if serverInfo["name"] != "vale-ste" {
		t.Errorf("serverInfo.name = %v, want vale-ste", serverInfo["name"])
	}
	if serverInfo["version"] != "1.2.3" {
		t.Errorf("serverInfo.version = %v, want 1.2.3", serverInfo["version"])
	}
	if initResult["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v", initResult["protocolVersion"])
	}

	// tools/list has 2 tools
	listResult := msgs[1]["result"].(map[string]any)
	tools := listResult["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("tools = %d, want 2", len(tools))
	}
	names := map[string]bool{}
	for _, tl := range tools {
		names[tl.(map[string]any)["name"].(string)] = true
	}
	if !names["lint_text"] || !names["list_rules"] {
		t.Errorf("tool names = %v, want lint_text and list_rules", names)
	}

	// tools/call returns content[0].text containing "findings"
	callResult := msgs[2]["result"].(map[string]any)
	content := callResult["content"].([]any)
	if len(content) == 0 {
		t.Fatal("empty content")
	}
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "findings") {
		t.Errorf("tool result text should contain 'findings': %q", text)
	}
	// The contraction should be reported.
	if !strings.Contains(text, "STE.Contractions") {
		t.Errorf("tool result should flag the contraction: %q", text)
	}
}

func TestServeIDsMatch(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n"
	msgs := runServer(t, input)
	if len(msgs) != 2 {
		t.Fatalf("responses = %d, want 2", len(msgs))
	}
	if msgs[0]["id"].(float64) != 1 || msgs[1]["id"].(float64) != 2 {
		t.Errorf("ids = %v, %v; want 1, 2", msgs[0]["id"], msgs[1]["id"])
	}
	for _, m := range msgs {
		if m["jsonrpc"] != "2.0" {
			t.Errorf("jsonrpc = %v", m["jsonrpc"])
		}
	}
}

func TestServeUnknownMethod(t *testing.T) {
	msgs := runServer(t, `{"jsonrpc":"2.0","id":9,"method":"no/such"}`+"\n")
	if len(msgs) != 1 {
		t.Fatalf("responses = %d, want 1", len(msgs))
	}
	errObj := msgs[0]["error"].(map[string]any)
	if errObj["code"].(float64) != -32601 {
		t.Errorf("error code = %v, want -32601", errObj["code"])
	}
}

func TestServeParseError(t *testing.T) {
	msgs := runServer(t, "{not valid json\n")
	if len(msgs) != 1 {
		t.Fatalf("responses = %d, want 1", len(msgs))
	}
	errObj := msgs[0]["error"].(map[string]any)
	if errObj["code"].(float64) != -32700 {
		t.Errorf("error code = %v, want -32700", errObj["code"])
	}
	// A parse error has a null id.
	if msgs[0]["id"] != nil {
		t.Errorf("id = %v, want null", msgs[0]["id"])
	}
}

func TestServeListRulesTool(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_rules","arguments":{}}}` + "\n"
	msgs := runServer(t, input)
	if len(msgs) != 1 {
		t.Fatalf("responses = %d, want 1", len(msgs))
	}
	result := msgs[0]["result"].(map[string]any)
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "7 rules") {
		t.Errorf("list_rules text should mention 7 rules: %q", text)
	}
	if !strings.Contains(text, "STE.Vocabulary") {
		t.Errorf("list_rules should include STE.Vocabulary: %q", text)
	}
}

func TestServeLintTextEmptyArg(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lint_text","arguments":{"text":""}}}` + "\n"
	msgs := runServer(t, input)
	result := msgs[0]["result"].(map[string]any)
	if result["isError"] != true {
		t.Errorf("empty text should be a tool error: %v", result)
	}
}

func TestServeLintTextMinSeverityFilter(t *testing.T) {
	// A contraction is an error; a min severity of error keeps it, and there
	// should be no suggestion-level findings in the output count.
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lint_text","arguments":{"text":"Don't stop.","minSeverity":"error"}}}` + "\n"
	msgs := runServer(t, input)
	result := msgs[0]["result"].(map[string]any)
	text := result["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "0 warnings, 0 suggestions") {
		t.Errorf("min severity error should drop warnings/suggestions: %q", text)
	}
}

func TestServeLintTextBadMinSeverity(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lint_text","arguments":{"text":"hi","minSeverity":"loud"}}}` + "\n"
	msgs := runServer(t, input)
	result := msgs[0]["result"].(map[string]any)
	if result["isError"] != true {
		t.Errorf("bad minSeverity should be a tool error: %v", result)
	}
}

func TestServeUnknownTool(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nope","arguments":{}}}` + "\n"
	msgs := runServer(t, input)
	errObj := msgs[0]["error"].(map[string]any)
	if errObj["code"].(float64) != -32602 {
		t.Errorf("unknown tool error code = %v, want -32602", errObj["code"])
	}
}

func TestServeSkipsBlankLines(t *testing.T) {
	input := "\n\n" + `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n\n"
	msgs := runServer(t, input)
	if len(msgs) != 1 {
		t.Fatalf("responses = %d, want 1 (blank lines skipped)", len(msgs))
	}
}

func TestNewServerDefaultVersion(t *testing.T) {
	l := linter.New(config.Default())
	s := NewServer(l, "")
	var out bytes.Buffer
	if err := s.Serve(context.Background(), strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`+"\n"), &out); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	line := strings.TrimSpace(out.String())
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatal(err)
	}
	info := m["result"].(map[string]any)["serverInfo"].(map[string]any)
	if info["version"] != "dev" {
		t.Errorf("default version = %v, want dev", info["version"])
	}
}

func TestServeContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	l := linter.New(config.Default())
	var out bytes.Buffer
	err := Serve(ctx, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n"), &out, l, "1")
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
