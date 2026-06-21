package gollmfree

import "context"

// Provider is the common contract implemented by every anonymous/free LLM
// provider. Implementations must respect context cancellation and avoid requiring
// caller-supplied credentials.
type Provider interface {
	// Name returns the stable registry name for this provider.
	Name() string

	// Complete returns one non-streaming chat completion for the supplied messages.
	Complete(ctx context.Context, messages []Message) (CompletionResponse, error)

	// Stream returns text chunks for the supplied messages and closes the channel
	// when complete. Providers without native streaming may emit one full-response
	// chunk.
	Stream(ctx context.Context, messages []Message) (<-chan string, error)

	// SupportedModels returns model aliases or provider names supported by this
	// provider. Returned aliases are routing hints, not authenticity guarantees.
	SupportedModels() []string
}

// ProviderInfo describes a provider registered with the client selector.
// Slices returned from registry APIs should be copies so callers cannot mutate
// shared registry state.
type ProviderInfo struct {
	Name            string
	Provider        Provider
	SupportedModels []string
	DefaultPriority int
}

// ModelInfo describes an alias and the providers that can handle it.
type ModelInfo struct {
	Alias     string
	Providers []string
}
