// Package lsp is a small Language Server Protocol server for vale. It gives
// editors live Simplified Technical English diagnostics: it lints each open
// document with the same engine as the CLI and the MCP server, and publishes the
// findings as LSP diagnostics. It speaks JSON-RPC 2.0 over stdio with
// Content-Length framing (LSP's transport, not the MCP newline framing).
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/linter"
)

// JSON-RPC error codes.
const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// LSP position and diagnostic types (0-based line and UTF-16 character).
type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type rangeT struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Range    rangeT `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// Server is an LSP server backed by a linter.
type Server struct {
	linter  *linter.Linter
	version string
	docs    map[string]string // uri -> text
	w       *bufio.Writer
	down    bool // shutdown requested
}

// NewServer builds a server from a linter.
func NewServer(l *linter.Linter, version string) *Server {
	if version == "" {
		version = "dev"
	}
	return &Server{linter: l, version: version, docs: map[string]string{}}
}

// Serve reads LSP messages from in and writes responses to out until exit or the
// end of input.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	r := bufio.NewReader(in)
	s.w = bufio.NewWriter(out)
	defer s.w.Flush()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		body, err := readMessage(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var req rpcRequest
		if json.Unmarshal(body, &req) != nil {
			s.respondError(nil, codeParseError, "parse error")
			continue
		}
		if req.Method == "exit" {
			return nil
		}
		s.handle(req)
	}
}

// handle dispatches one message.
func (s *Server) handle(req rpcRequest) {
	switch req.Method {
	case "initialize":
		s.respond(req.ID, s.initializeResult())
	case "initialized", "$/setTrace", "$/cancelRequest":
		// notifications with no effect
	case "shutdown":
		s.down = true
		s.respond(req.ID, nil)
	case "textDocument/didOpen":
		s.didOpen(req.Params)
	case "textDocument/didChange":
		s.didChange(req.Params)
	case "textDocument/didSave":
		s.didSave(req.Params)
	case "textDocument/didClose":
		s.didClose(req.Params)
	default:
		if len(req.ID) > 0 { // a request needs a response
			s.respondError(req.ID, codeMethodNotFound, "unknown method: "+req.Method)
		}
	}
}

// initializeResult advertises full-sync text documents and push diagnostics.
func (s *Server) initializeResult() any {
	return map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": 1, // full document sync
			"diagnosticProvider": map[string]any{
				"interFileDependencies": false,
				"workspaceDiagnostics":  false,
			},
		},
		"serverInfo": map[string]any{"name": "vale-ste", "version": s.version},
	}
}

// didOpenParams and friends are the subset of LSP params the server reads.
type textDoc struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

func (s *Server) didOpen(raw json.RawMessage) {
	var p struct {
		TextDocument textDoc `json:"textDocument"`
	}
	if json.Unmarshal(raw, &p) != nil {
		return
	}
	s.docs[p.TextDocument.URI] = p.TextDocument.Text
	s.publish(p.TextDocument.URI)
}

func (s *Server) didChange(raw json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		ContentChanges []struct {
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	if json.Unmarshal(raw, &p) != nil || len(p.ContentChanges) == 0 {
		return
	}
	// Full sync: the last change carries the whole document.
	s.docs[p.TextDocument.URI] = p.ContentChanges[len(p.ContentChanges)-1].Text
	s.publish(p.TextDocument.URI)
}

func (s *Server) didSave(raw json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if json.Unmarshal(raw, &p) == nil {
		s.publish(p.TextDocument.URI)
	}
}

func (s *Server) didClose(raw json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if json.Unmarshal(raw, &p) != nil {
		return
	}
	delete(s.docs, p.TextDocument.URI)
	// Clear diagnostics for the closed document.
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri": p.TextDocument.URI, "diagnostics": []diagnostic{},
	})
}

// publish lints the document and sends its diagnostics.
func (s *Server) publish(uri string) {
	text, ok := s.docs[uri]
	if !ok {
		return
	}
	findings := s.linter.LintText(filenameFromURI(uri), text, linter.MarkdownAuto)
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": toDiagnostics(text, findings),
	})
}

// toDiagnostics converts findings to LSP diagnostics with 0-based, UTF-16
// positions.
func toDiagnostics(text string, findings []lint.Finding) []diagnostic {
	lines := strings.Split(text, "\n")
	out := make([]diagnostic, 0, len(findings))
	for _, f := range findings {
		msg := f.Message
		if f.Hint != "" {
			msg += " — " + f.Hint
		}
		out = append(out, diagnostic{
			Range: rangeT{
				Start: position{Line: max0(f.Line - 1), Character: utf16Col(lineAt(lines, f.Line), f.Col)},
				End:   position{Line: max0(f.EndLine - 1), Character: utf16Col(lineAt(lines, f.EndLine), f.EndCol)},
			},
			Severity: severity(f.Severity),
			Source:   "vale",
			Code:     f.RuleID,
			Message:  msg,
		})
	}
	return out
}

// severity maps an STE severity to an LSP DiagnosticSeverity.
func severity(s lint.Severity) int {
	switch s {
	case lint.SeverityError:
		return 1 // Error
	case lint.SeverityWarning:
		return 2 // Warning
	default:
		return 3 // Information
	}
}

// utf16Col converts a 1-based rune column to a 0-based UTF-16 character offset.
func utf16Col(line string, col1 int) int {
	if col1 <= 1 {
		return 0
	}
	runes := []rune(line)
	n := col1 - 1
	if n > len(runes) {
		n = len(runes)
	}
	units := 0
	for _, r := range runes[:n] {
		if r > 0xFFFF {
			units += 2
		} else {
			units++
		}
	}
	return units
}

func lineAt(lines []string, line1 int) string {
	if line1 >= 1 && line1 <= len(lines) {
		return lines[line1-1]
	}
	return ""
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// filenameFromURI turns a file:// URI into a path for the Markdown decision.
func filenameFromURI(uri string) string {
	if p, ok := strings.CutPrefix(uri, "file://"); ok {
		return p
	}
	return uri
}

// respond writes a JSON-RPC result.
func (s *Server) respond(id json.RawMessage, result any) {
	_ = writeMessage(s.w, rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

// respondError writes a JSON-RPC error.
func (s *Server) respondError(id json.RawMessage, code int, msg string) {
	_ = writeMessage(s.w, rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

// notify sends a server-to-client notification.
func (s *Server) notify(method string, params any) {
	_ = writeMessage(s.w, notification{JSONRPC: "2.0", Method: method, Params: params})
}

// readMessage reads one Content-Length framed message body.
func readMessage(r *bufio.Reader) ([]byte, error) {
	length := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end of headers
		}
		if rest, ok := cutHeader(line, "content-length:"); ok {
			n, err := strconv.Atoi(strings.TrimSpace(rest))
			if err != nil {
				return nil, err
			}
			length = n
		}
	}
	if length <= 0 {
		return nil, fmt.Errorf("missing or invalid Content-Length")
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// cutHeader matches a header name case-insensitively and returns its value.
func cutHeader(line, name string) (string, bool) {
	if len(line) >= len(name) && strings.EqualFold(line[:len(name)], name) {
		return line[len(name):], true
	}
	return "", false
}

// writeMessage writes one Content-Length framed JSON message.
func writeMessage(w *bufio.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return w.Flush()
}

// Serve is a helper that builds a server and serves once.
func Serve(ctx context.Context, in io.Reader, out io.Writer, l *linter.Linter, version string) error {
	return NewServer(l, version).Serve(ctx, in, out)
}
