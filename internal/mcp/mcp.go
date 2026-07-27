// Package mcp is a small Model Context Protocol server. It speaks JSON-RPC 2.0
// over standard input and output, with one JSON message on each line. It
// exposes the Simplified Technical English engine as MCP tools, so an agent can
// lint text through the same engine as the command line.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/stuffbucket/vale/internal/linter"
)

// protocolVersion is the MCP version that the server reports when the client
// does not ask for one.
const protocolVersion = "2024-11-05"

// serverName and serverVersion identify the server to the client.
const serverName = "vale-ste"

// JSON-RPC error codes from the specification.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// request is one JSON-RPC request or notification. A notification has no id.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// response is one JSON-RPC response.
type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcError is a JSON-RPC error object.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Server holds the linter and the server version.
type Server struct {
	linter  *linter.Linter
	version string
}

// NewServer builds a server that uses the given linter. The version is for the
// serverInfo that the initialize response returns.
func NewServer(l *linter.Linter, version string) *Server {
	if version == "" {
		version = "dev"
	}
	return &Server{linter: l, version: version}
}

// Serve reads requests from in and writes responses to out until in ends. It
// returns nil at a clean end of input.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		resp, send := s.handleLine(line)
		if !send {
			continue
		}
		if err := writeMessage(writer, resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// handleLine handles one line of input. It returns the response and whether to
// send it. Notifications produce no response.
func (s *Server) handleLine(line []byte) (response, bool) {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return errorResponse(nil, codeParseError, "parse error"), true
	}
	isNotification := len(req.ID) == 0
	result, rpcErr := s.dispatch(req)
	if isNotification {
		return response{}, false
	}
	if rpcErr != nil {
		return response{JSONRPC: "2.0", ID: req.ID, Error: rpcErr}, true
	}
	return response{JSONRPC: "2.0", ID: req.ID, Result: result}, true
}

// dispatch runs the method and returns the result or an error.
func (s *Server) dispatch(req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return s.initialize(req.Params)
	case "notifications/initialized":
		return nil, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return s.toolsList(), nil
	case "tools/call":
		return s.toolsCall(req.Params)
	default:
		return nil, &rpcError{Code: codeMethodNotFound, Message: "unknown method: " + req.Method}
	}
}

// initialize answers the initialize request.
func (s *Server) initialize(params json.RawMessage) (any, *rpcError) {
	version := protocolVersion
	if len(params) > 0 {
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if err := json.Unmarshal(params, &p); err == nil && p.ProtocolVersion != "" {
			version = p.ProtocolVersion
		}
	}
	return map[string]any{
		"protocolVersion": version,
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    serverName,
			"version": s.version,
		},
	}, nil
}

// writeMessage writes one JSON message and a newline.
func writeMessage(w *bufio.Writer, msg response) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}

// errorResponse builds an error response.
func errorResponse(id json.RawMessage, code int, message string) response {
	return response{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
}

// Serve is a helper that builds a server and serves once.
func Serve(ctx context.Context, in io.Reader, out io.Writer, l *linter.Linter, version string) error {
	return NewServer(l, version).Serve(ctx, in, out)
}
