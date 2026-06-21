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
	chataiName         = "chatai"
	chataiDefaultModel = "gpt-4o-mini"
	chataiEndpoint     = "https://chatai.ren/api/chat"
	// Static anonymous token expected by the Chatai mobile API.
	// Upstream source: xtekky/gpt4free g4f/Provider/Chatai.py at 798d8586.
	chataiStaticToken = "ddnon5svtivajdspg7rscqw5"
)

// Chatai is a no-auth SSE provider backed by chatai.ren.
// Upstream: xtekky/gpt4free g4f/Provider/Chatai.py at 798d8586;
// working=True, needs_auth=False, supports_stream=True.
type Chatai struct {
	client   *http.Client
	endpoint string
	token    string
}

// ChataiOption configures Chatai.
type ChataiOption func(*Chatai)

// NewChatai constructs a Chatai provider.
func NewChatai(opts ...ChataiOption) *Chatai {
	p := &Chatai{client: http.DefaultClient, endpoint: chataiEndpoint, token: chataiStaticToken}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// WithChataiEndpoint overrides the API endpoint for tests.
func WithChataiEndpoint(endpoint string) ChataiOption {
	return func(p *Chatai) {
		if endpoint != "" {
			p.endpoint = endpoint
		}
	}
}

// WithChataiHTTPClient overrides the HTTP client.
func WithChataiHTTPClient(client *http.Client) ChataiOption {
	return func(p *Chatai) {
		if client != nil {
			p.client = client
		}
	}
}

// WithChataiToken overrides the static anonymous token.
func WithChataiToken(token string) ChataiOption {
	return func(p *Chatai) { p.token = token }
}

// Name implements Provider.
func (p *Chatai) Name() string { return chataiName }

// SupportedModels implements Provider.
func (p *Chatai) SupportedModels() []string {
	return []string{"chatai", chataiDefaultModel, "gpt-4o-mini-chatai"}
}

// Complete sends a chat request and returns the accumulated SSE response.
func (p *Chatai) Complete(ctx context.Context, messages []gollmfree.Message) (gollmfree.CompletionResponse, error) {
	body, err := json.Marshal(chataiRequest{
		MachineID: "gollmfree",
		Msg:       messages,
		Token:     p.token,
		Type:      1,
	})
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: marshal request: %w", p.Name(), err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: create request: %w", p.Name(), err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

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

	text, err := chataiReadSSE(resp.Body)
	if err != nil {
		return gollmfree.CompletionResponse{}, fmt.Errorf("%s: read SSE: %w", p.Name(), err)
	}
	return gollmfree.CompletionResponse{
		Model:    chataiDefaultModel,
		Provider: p.Name(),
		Choices: []gollmfree.Choice{{
			Message:      gollmfree.Message{Role: "assistant", Content: text},
			FinishReason: "stop",
		}},
	}, nil
}

// Stream implements Provider using SSE streaming.
func (p *Chatai) Stream(ctx context.Context, messages []gollmfree.Message) (<-chan string, error) {
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

// chataiReadSSE accumulates content from OpenAI-delta SSE lines.
func chataiReadSSE(r io.Reader) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}
		var chunk chataiDelta
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			sb.WriteString(chunk.Choices[0].Delta.Content)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

type chataiRequest struct {
	MachineID string              `json:"machineId"`
	Msg       []gollmfree.Message `json:"msg"`
	Token     string              `json:"token"`
	Type      int                 `json:"type"`
}

type chataiDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
