package providers

import (
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
	pollinationsName         = "pollinationsai"
	pollinationsDefaultModel = "openai-fast"
	pollinationsTextEndpoint = "https://text.pollinations.ai/openai"
)

// PollinationsAI implements the current upstream no-auth Pollinations text path.
type PollinationsAI struct {
	client   *http.Client
	endpoint string
}

// PollinationsOption configures PollinationsAI.
type PollinationsOption func(*PollinationsAI)

// NewPollinationsAI constructs a PollinationsAI provider.
func NewPollinationsAI(opts ...PollinationsOption) *PollinationsAI {
	provider := &PollinationsAI{client: http.DefaultClient, endpoint: pollinationsTextEndpoint}
	for _, opt := range opts {
		if opt != nil {
			opt(provider)
		}
	}
	return provider
}

// WithPollinationsEndpoint overrides the OpenAI-shaped text endpoint for tests.
func WithPollinationsEndpoint(endpoint string) PollinationsOption {
	return func(provider *PollinationsAI) {
		if endpoint != "" {
			provider.endpoint = endpoint
		}
	}
}

// WithPollinationsHTTPClient overrides the HTTP client for tests or embedding.
func WithPollinationsHTTPClient(client *http.Client) PollinationsOption {
	return func(provider *PollinationsAI) {
		if client != nil {
			provider.client = client
		}
	}
}

// Name returns the registry name for PollinationsAI.
func (p *PollinationsAI) Name() string { return pollinationsName }

// SupportedModels returns aliases routed to PollinationsAI for v0.1.0.
func (p *PollinationsAI) SupportedModels() []string {
	return []string{"auto", "best", "pollinationsai", pollinationsDefaultModel, "gpt-4.1-nano"}
}

// Complete sends a non-streaming OpenAI-shaped chat request to PollinationsAI.
func (p *PollinationsAI) Complete(ctx context.Context, messages []gollmfree.Message) (gollmfree.CompletionResponse, error) {
	payload := pollinationsChatRequest{Model: pollinationsDefaultModel, Messages: messages, Stream: false}
	body, err := json.Marshal(payload)
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: marshal request: %w", p.Name(), err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: create request: %w", p.Name(), err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: post completion: %w", p.Name(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if snippet := strings.TrimSpace(string(body)); snippet != "" {
			return gollmfree.CompletionResponse{}, fmt.Errorf("%s: unexpected status %s: %s", p.Name(), resp.Status, snippet)
		}
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: unexpected status %s", p.Name(), resp.Status)
	}

	var completion gollmfree.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: decode response: %w", p.Name(), err)
	}
	completion.Provider = p.Name()
	if completion.Model == "" {
		completion.Model = pollinationsDefaultModel
	}
	return completion, nil
}

// Stream emulates streaming by returning the complete response as one chunk.
func (p *PollinationsAI) Stream(ctx context.Context, messages []gollmfree.Message) (<-chan string, error) {
	completion, err := p.Complete(ctx, messages)
	if err != nil {
		return nil, err
	}
	chunks := make(chan string, 1)
	if len(completion.Choices) > 0 {
		chunks <- completion.Choices[0].Message.Content
	}
	close(chunks)
	return chunks, nil
}

type pollinationsChatRequest struct {
	Model    string              `json:"model"`
	Messages []gollmfree.Message `json:"messages"`
	Stream   bool                `json:"stream"`
}
