package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

const (
	yqcloudName         = "yqcloud"
	yqcloudDefaultModel = "gpt-3.5-turbo"
	yqcloudEndpoint     = "https://api.aichatos.cloud/api/generateStream"
)

// Yqcloud is a no-auth plain-stream provider backed by the Yqcloud API.
// Upstream: xtekky/gpt4free g4f/Provider/Yqcloud.py at 798d8586;
// working=True, needs_auth=False, plain-text stream per line.
type Yqcloud struct {
	client   *http.Client
	endpoint string
}

// YqcloudOption configures Yqcloud.
type YqcloudOption func(*Yqcloud)

// NewYqcloud constructs a Yqcloud provider.
func NewYqcloud(opts ...YqcloudOption) *Yqcloud {
	p := &Yqcloud{client: http.DefaultClient, endpoint: yqcloudEndpoint}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// WithYqcloudEndpoint overrides the API endpoint for tests.
func WithYqcloudEndpoint(endpoint string) YqcloudOption {
	return func(p *Yqcloud) {
		if endpoint != "" {
			p.endpoint = endpoint
		}
	}
}

// WithYqcloudHTTPClient overrides the HTTP client.
func WithYqcloudHTTPClient(client *http.Client) YqcloudOption {
	return func(p *Yqcloud) {
		if client != nil {
			p.client = client
		}
	}
}

// Name implements Provider.
func (p *Yqcloud) Name() string { return yqcloudName }

// SupportedModels implements Provider.
func (p *Yqcloud) SupportedModels() []string {
	return []string{"yqcloud", "gpt-3.5-turbo-yqcloud"}
}

// Complete sends a chat request and accumulates the plain-text stream response.
func (p *Yqcloud) Complete(ctx context.Context, messages []gollmfree.Message) (gollmfree.CompletionResponse, error) {
	prompt := yqcloudBuildPrompt(messages)
	body, err := json.Marshal(yqcloudRequest{
		Prompt:   prompt,
		UserID:   "gollmfree",
		Network:  true,
		System:   "",
		WithGPT:  false,
	})
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: marshal request: %w", p.Name(), err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: create request: %w", p.Name(), err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://chat9.yqcloud.top")
	req.Header.Set("Referer", "https://chat9.yqcloud.top/")

	resp, err := p.client.Do(req)
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: post: %w", p.Name(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if s := strings.TrimSpace(string(snippet)); s != "" {
			return gollmfree.CompletionResponse{}, fmt.Errorf("%s: unexpected status %s: %s", p.Name(), resp.Status, s)
		}
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: unexpected status %s", p.Name(), resp.Status)
	}

	text, err := yqcloudReadStream(resp.Body)
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: read stream: %w", p.Name(), err)
	}
	return gollmfree.CompletionResponse{
		Model:    yqcloudDefaultModel,
		Provider: p.Name(),
		Choices: []gollmfree.Choice{{
			Message:      gollmfree.Message{Role: "assistant", Content: text},
			FinishReason: "stop",
		}},
	}, nil
}

// Stream implements Provider with emulated single-chunk streaming.
func (p *Yqcloud) Stream(ctx context.Context, messages []gollmfree.Message) (<-chan string, error) {
	resp, err := p.Complete(ctx, messages)
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 1)
	if len(resp.Choices) > 0 {
		ch <- resp.Choices[0].Message.Content
	}
	close(ch)
	return ch, nil
}

// yqcloudReadStream reads a plain-text stream where each non-empty line is a content chunk.
func yqcloudReadStream(r io.Reader) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		sb.WriteString(line)
	}
	return sb.String(), scanner.Err()
}

// yqcloudBuildPrompt flattens the message list into a single prompt string.
// The Yqcloud API takes a simple prompt rather than a structured messages array.
func yqcloudBuildPrompt(messages []gollmfree.Message) string {
	if len(messages) == 0 {
		return ""
	}
	// Use the last user message as the prompt to keep it simple.
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return messages[len(messages)-1].Content
}

type yqcloudRequest struct {
	Prompt  string `json:"prompt"`
	UserID  string `json:"userId"`
	Network bool   `json:"network"`
	System  string `json:"system"`
	WithGPT bool   `json:"withGPT"`
}
