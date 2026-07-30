package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
)

// frame wraps each JSON message in Content-Length framing.
func frame(msgs ...string) string {
	var b strings.Builder
	for _, m := range msgs {
		fmt.Fprintf(&b, "Content-Length: %d\r\n\r\n%s", len(m), m)
	}
	return b.String()
}

// runLSP serves the framed input and returns the parsed output messages.
func runLSP(t *testing.T, input string) []map[string]any {
	t.Helper()
	l := linter.New(config.Default())
	var out bytes.Buffer
	if err := Serve(context.Background(), strings.NewReader(input), &out, l, "1.0"); err != nil {
		t.Fatalf("serve: %v", err)
	}
	return parseFrames(t, out.String())
}

// parseFrames decodes Content-Length framed JSON messages.
func parseFrames(t *testing.T, s string) []map[string]any {
	t.Helper()
	var msgs []map[string]any
	r := bufio.NewReader(strings.NewReader(s))
	for {
		length := 0
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return msgs
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			if strings.HasPrefix(strings.ToLower(line), "content-length:") {
				length, _ = strconv.Atoi(strings.TrimSpace(line[len("content-length:"):]))
			}
		}
		if length == 0 {
			return msgs
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return msgs
		}
		var m map[string]any
		if err := json.Unmarshal(buf, &m); err != nil {
			t.Fatalf("bad frame: %v", err)
		}
		msgs = append(msgs, m)
	}
}

func TestLSPInitializeAndDiagnostics(t *testing.T) {
	input := frame(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"initialized"}`,
		`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///x.md","text":"Don't stop."}}}`,
	)
	msgs := runLSP(t, input)
	if len(msgs) < 2 {
		t.Fatalf("expected an initialize result and a diagnostics notification: %v", msgs)
	}

	// initialize result
	init := msgs[0]["result"].(map[string]any)
	caps := init["capabilities"].(map[string]any)
	if caps["textDocumentSync"].(float64) != 1 {
		t.Errorf("textDocumentSync = %v, want 1", caps["textDocumentSync"])
	}
	if init["serverInfo"].(map[string]any)["name"] != "vale-ste" {
		t.Errorf("serverInfo.name = %v", init["serverInfo"])
	}

	// publishDiagnostics notification
	var diag map[string]any
	for _, m := range msgs {
		if m["method"] == "textDocument/publishDiagnostics" {
			diag = m
			break
		}
	}
	if diag == nil {
		t.Fatal("no publishDiagnostics notification")
	}
	params := diag["params"].(map[string]any)
	if params["uri"] != "file:///x.md" {
		t.Errorf("uri = %v", params["uri"])
	}
	diags := params["diagnostics"].([]any)
	if len(diags) == 0 {
		t.Fatal("a contraction should produce a diagnostic")
	}
	d0 := diags[0].(map[string]any)
	if d0["code"] != "STE.Contractions" {
		t.Errorf("code = %v, want STE.Contractions", d0["code"])
	}
	if d0["severity"].(float64) != 1 {
		t.Errorf("severity = %v, want 1 (Error)", d0["severity"])
	}
	if d0["source"] != "vale" {
		t.Errorf("source = %v, want vale", d0["source"])
	}
	start := d0["range"].(map[string]any)["start"].(map[string]any)
	if start["line"].(float64) != 0 || start["character"].(float64) != 0 {
		t.Errorf("start = %v, want line 0 char 0", start)
	}
}

func TestLSPDidCloseClearsDiagnostics(t *testing.T) {
	input := frame(
		`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///x.md","text":"Don't stop."}}}`,
		`{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"file:///x.md"}}}`,
	)
	msgs := runLSP(t, input)
	// The last publishDiagnostics for the uri should be empty.
	var last map[string]any
	for _, m := range msgs {
		if m["method"] == "textDocument/publishDiagnostics" {
			last = m
		}
	}
	if last == nil {
		t.Fatal("no diagnostics messages")
	}
	if d := last["params"].(map[string]any)["diagnostics"].([]any); len(d) != 0 {
		t.Errorf("close should clear diagnostics, got %d", len(d))
	}
}

func TestLSPUnknownRequestErrors(t *testing.T) {
	msgs := runLSP(t, frame(`{"jsonrpc":"2.0","id":9,"method":"no/such"}`))
	if len(msgs) != 1 {
		t.Fatalf("responses = %d, want 1", len(msgs))
	}
	if msgs[0]["error"].(map[string]any)["code"].(float64) != codeMethodNotFound {
		t.Errorf("want method-not-found, got %v", msgs[0]["error"])
	}
}

func TestUTF16Col(t *testing.T) {
	if got := utf16Col("hello", 1); got != 0 {
		t.Errorf("col 1 = %d, want 0", got)
	}
	if got := utf16Col("hello", 3); got != 2 {
		t.Errorf("ascii col 3 = %d, want 2", got)
	}
	// An astral character (emoji) is two UTF-16 code units.
	if got := utf16Col("😀x", 2); got != 2 {
		t.Errorf("prefix past emoji = %d, want 2", got)
	}
}

func TestReadMessageFraming(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	framed := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
	got, err := readMessage(bufio.NewReader(strings.NewReader(framed)))
	if err != nil || string(got) != body {
		t.Fatalf("readMessage = %q, err %v", got, err)
	}
}
