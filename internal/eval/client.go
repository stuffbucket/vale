// Package eval drives a configured OpenAI-compatible LLM endpoint (by default
// the local proxy at http://localhost:4141, which serves several model
// families), lints each model's output with vale, and reports how much "slop"
// each model and family produces. It turns the STE.Slop* research into a
// measurement.
package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is the local multi-family proxy.
const DefaultEndpoint = "http://localhost:4141"

// Client talks to an OpenAI-compatible chat endpoint.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// NewClient builds a client with a sane default timeout.
func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = DefaultEndpoint
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// Model is one available model and its owning family.
type Model struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

// Models lists the models the endpoint offers (GET /v1/models).
func (c *Client) Models(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	var body struct {
		Data []Model `json:"data"`
	}
	if err := c.do(req, &body); err != nil {
		return nil, err
	}
	return body.Data, nil
}

// Message is one chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Complete sends a single-turn prompt and returns the reply text. It tries the
// chat-completions API first and transparently falls back to the Responses API
// for models the endpoint only serves there. A non-nil temperature is passed to
// the model; nil uses the server default.
func (c *Client) Complete(ctx context.Context, model, prompt string, maxTokens int, temperature *float64) (string, error) {
	out, err := c.completeChat(ctx, model, prompt, maxTokens, temperature)
	if err != nil && unsupportedChatAPI(err) {
		return c.completeResponses(ctx, model, prompt, maxTokens, temperature)
	}
	return out, err
}

// unsupportedChatAPI reports whether the error means the model needs the
// Responses API instead of chat completions.
func unsupportedChatAPI(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not accessible via the /chat/completions") ||
		strings.Contains(msg, "unsupported_api_for_model")
}

// completeChat uses POST /v1/chat/completions.
func (c *Client) completeChat(ctx context.Context, model, prompt string, maxTokens int, temperature *float64) (string, error) {
	reqBody := map[string]any{
		"model":    model,
		"messages": []Message{{Role: "user", Content: prompt}},
	}
	if maxTokens > 0 {
		reqBody["max_tokens"] = maxTokens
	}
	if temperature != nil {
		reqBody["temperature"] = *temperature
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.auth(req)
	var body struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := c.do(req, &body); err != nil {
		return "", err
	}
	if len(body.Choices) == 0 {
		return "", fmt.Errorf("no choices in response from %s", model)
	}
	return body.Choices[0].Message.Content, nil
}

// completeResponses uses POST /v1/responses and extracts the message text from
// the output items (skipping reasoning items).
func (c *Client) completeResponses(ctx context.Context, model, prompt string, maxTokens int, temperature *float64) (string, error) {
	reqBody := map[string]any{"model": model, "input": prompt}
	if maxTokens > 0 {
		reqBody["max_output_tokens"] = maxTokens
	}
	if temperature != nil {
		reqBody["temperature"] = *temperature
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/responses", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.auth(req)
	var body struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := c.do(req, &body); err != nil {
		return "", err
	}
	if body.OutputText != "" {
		return body.OutputText, nil
	}
	var sb strings.Builder
	for _, item := range body.Output {
		if item.Type != "message" {
			continue
		}
		for _, part := range item.Content {
			if part.Type == "output_text" {
				sb.WriteString(part.Text)
			}
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("empty responses output from %s", model)
	}
	return sb.String(), nil
}

// auth adds the bearer token when one is set.
func (c *Client) auth(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
}

// do runs a request and decodes a JSON response, surfacing non-2xx bodies.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s: HTTP %d: %s", req.URL.Path, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return json.Unmarshal(data, out)
}
