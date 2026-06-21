package gollmfree

// Message is one chat message exchanged with a provider.
// Role is typically "system", "user", or "assistant"; Content holds the
// natural-language text sent to or returned by the model.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest describes a chat-completion request in an OpenAI-shaped form.
// Model is a routing hint such as "auto", a provider name, or a provider-claimed
// model alias. Messages are sent to the selected provider in order.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
}

// CompletionResponse is a provider response normalized into an OpenAI-shaped
// chat-completion payload. Provider identifies the actual provider used when it
// is known.
type CompletionResponse struct {
	ID       string   `json:"id,omitempty"`
	Object   string   `json:"object,omitempty"`
	Created  int64    `json:"created,omitempty"`
	Model    string   `json:"model,omitempty"`
	Provider string   `json:"provider,omitempty"`
	Choices  []Choice `json:"choices"`
}

// Choice is one completion alternative returned by a provider.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// StreamChunk is one text fragment returned by ChatCompletionStream.
// Providers without native streaming may emit a single chunk containing the full
// completion text.
type StreamChunk struct {
	Content  string `json:"content"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}
