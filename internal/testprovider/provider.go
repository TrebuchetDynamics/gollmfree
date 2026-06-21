package testprovider

import (
	"context"
	"sync/atomic"
	"time"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

// Provider is a deterministic fake gollmfree.Provider for unit tests.
type Provider struct {
	name         string
	models       []string
	responseText string
	streamChunks []string
	delay        time.Duration
	err          error
	completeN    atomic.Int64
	streamN      atomic.Int64
}

// Option configures a fake Provider.
type Option func(*Provider)

// New returns a fake provider with the supplied name.
func New(name string, opts ...Option) *Provider {
	provider := &Provider{name: name, models: []string{"auto"}, responseText: "ok"}
	for _, opt := range opts {
		if opt != nil {
			opt(provider)
		}
	}
	return provider
}

// WithModels sets the aliases returned by SupportedModels.
func WithModels(models ...string) Option {
	return func(provider *Provider) {
		provider.models = append([]string(nil), models...)
	}
}

// WithResponse sets the assistant text returned by Complete.
func WithResponse(text string) Option {
	return func(provider *Provider) {
		provider.responseText = text
	}
}

// WithStreamChunks sets chunks emitted by Stream.
func WithStreamChunks(chunks ...string) Option {
	return func(provider *Provider) {
		provider.streamChunks = append([]string(nil), chunks...)
	}
}

// WithDelay makes Complete and Stream wait before returning unless the context
// is canceled first.
func WithDelay(delay time.Duration) Option {
	return func(provider *Provider) {
		provider.delay = delay
	}
}

// WithError makes Complete and Stream return err.
func WithError(err error) Option {
	return func(provider *Provider) {
		provider.err = err
	}
}

// Name returns the fake provider name.
func (p *Provider) Name() string { return p.name }

// SupportedModels returns a copy of configured aliases.
func (p *Provider) SupportedModels() []string {
	return append([]string(nil), p.models...)
}

// Complete records a call and returns the configured response or error.
func (p *Provider) Complete(ctx context.Context, messages []gollmfree.Message) (gollmfree.CompletionResponse, error) {
	p.completeN.Add(1)
	if err := p.wait(ctx); err != nil {
		return gollmfree.CompletionResponse{}, err
	}
	if p.err != nil {
		return gollmfree.CompletionResponse{}, p.err
	}
	return gollmfree.CompletionResponse{
		Provider: p.name,
		Choices:  []gollmfree.Choice{{Message: gollmfree.Message{Role: "assistant", Content: p.responseText}}},
	}, nil
}

// Stream records a call and returns the configured chunks or error.
func (p *Provider) Stream(ctx context.Context, messages []gollmfree.Message) (<-chan string, error) {
	p.streamN.Add(1)
	if err := p.wait(ctx); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	chunks := p.streamChunks
	if len(chunks) == 0 {
		chunks = []string{p.responseText}
	}
	ch := make(chan string, len(chunks))
	for _, chunk := range chunks {
		ch <- chunk
	}
	close(ch)
	return ch, nil
}

// CompleteCalls returns how many times Complete was called.
func (p *Provider) CompleteCalls() int64 { return p.completeN.Load() }

// StreamCalls returns how many times Stream was called.
func (p *Provider) StreamCalls() int64 { return p.streamN.Load() }

func (p *Provider) wait(ctx context.Context) error {
	if p.delay <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(p.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
