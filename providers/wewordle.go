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
	wewordleName         = "wewordle"
	wewordleDefaultModel = "gpt-3.5-turbo"
	wewordleEndpoint     = "https://wewordle.org/gptapi/v1/en/trial"
)

// WeWordle is a no-auth JSON-completion provider backed by wewordle.org.
// Upstream: xtekky/gpt4free g4f/Provider/WeWordle.py at 798d8586;
// working=True, needs_auth=False, non-streaming JSON response.
type WeWordle struct {
	client   *http.Client
	endpoint string
}

// WeWordleOption configures WeWordle.
type WeWordleOption func(*WeWordle)

// NewWeWordle constructs a WeWordle provider.
func NewWeWordle(opts ...WeWordleOption) *WeWordle {
	p := &WeWordle{client: http.DefaultClient, endpoint: wewordleEndpoint}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// WithWeWordleEndpoint overrides the API endpoint for tests.
func WithWeWordleEndpoint(endpoint string) WeWordleOption {
	return func(p *WeWordle) {
		if endpoint != "" {
			p.endpoint = endpoint
		}
	}
}

// WithWeWordleHTTPClient overrides the HTTP client.
func WithWeWordleHTTPClient(client *http.Client) WeWordleOption {
	return func(p *WeWordle) {
		if client != nil {
			p.client = client
		}
	}
}

// Name implements Provider.
func (p *WeWordle) Name() string { return wewordleName }

// SupportedModels implements Provider.
func (p *WeWordle) SupportedModels() []string {
	return []string{"wewordle", "gpt-3.5-turbo-wewordle"}
}

// Complete sends a chat request and returns the JSON response content.
func (p *WeWordle) Complete(ctx context.Context, messages []gollmfree.Message) (gollmfree.CompletionResponse, error) {
	contents := make([]wewordleContent, len(messages))
	for i, m := range messages {
		contents[i] = wewordleContent{Role: m.Role, Content: m.Content}
	}
	body, err := json.Marshal(wewordleRequest{
		Subscriber: wewordleSubscriber{},
		App:        wewordleApp{},
		User:       wewordleUser{FingerPrint: nil},
		Messages:   contents,
	})
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

	var result wewordleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: decode response: %w", p.Name(), err)
	}
	return gollmfree.CompletionResponse{
		Model:    wewordleDefaultModel,
		Provider: p.Name(),
		Choices: []gollmfree.Choice{{
			Message:      gollmfree.Message{Role: "assistant", Content: result.Content},
			FinishReason: "stop",
		}},
	}, nil
}

// Stream implements Provider with emulated single-chunk streaming.
func (p *WeWordle) Stream(ctx context.Context, messages []gollmfree.Message) (<-chan string, error) {
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

type wewordleSubscriber struct {
	OriginalPurchaseDate        *string `json:"originalPurchaseDate"`
	OriginalApplicationVersion  *string `json:"originalApplicationVersion"`
}

type wewordleApp struct {
	Version *string `json:"version"`
	Build   *string `json:"build"`
}

type wewordleUser struct {
	FingerPrint *string `json:"fingerPrint"`
}

type wewordleContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type wewordleRequest struct {
	Subscriber wewordleSubscriber `json:"subscriber"`
	App        wewordleApp        `json:"app"`
	User       wewordleUser       `json:"user"`
	Messages   []wewordleContent  `json:"messages"`
}

type wewordleResponse struct {
	Content string `json:"content"`
}
