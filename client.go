package gollmfree

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client is the concurrency-safe public entry point for chat completions.
// Construction does not perform network calls.
type Client struct {
	registry *Registry
	selector *Selector
	health   *HealthStore
	http     *http.Client
	options  clientOptions
}

// NewClient constructs a client with default registry, HTTP client, and options.
func NewClient(opts ...Option) *Client {
	options := defaultClientOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	registry := options.registryOverride
	if registry == nil {
		var err error
		registry, err = NewRegistry()
		if err != nil {
			registry = &Registry{}
		}
	}
	return &Client{
		registry: registry,
		selector: NewSelector(),
		health:   NewHealthStore(3, 5*time.Minute),
		http:     http.DefaultClient,
		options:  options,
	}
}

// ChatCompletion attempts ranked providers until one returns a completion.
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (CompletionResponse, error) {
	if c == nil {
		return CompletionResponse{}, fmt.Errorf("gollmfree: nil client")
	}
	model := normalizeRegistryKey(req.Model)
	if model == "" {
		model = c.options.defaultModel
	}
	candidates := c.registry.Candidates(model)
	if len(candidates) == 0 {
		return CompletionResponse{}, fmt.Errorf("gollmfree: no providers for model %q", model)
	}
	ranked := c.selector.Rank(model, candidates, c.health.Snapshot(), c.options)
	if c.options.raceMode && len(ranked) > 1 {
		resp, attempts, ok := c.raceCompletion(ctx, model, req, ranked)
		if ok {
			return resp, nil
		}
		if len(attempts) > 0 {
			return CompletionResponse{}, CombinedError{Attempts: attempts}
		}
	}
	attempts := make([]AttemptError, 0, len(ranked))
	for _, candidate := range ranked {
		if candidate.Provider == nil {
			continue
		}
		for attempt := 0; attempt <= c.options.maxRetries; attempt++ {
			attemptCtx, cancel := context.WithTimeout(ctx, c.options.perAttemptTimeout)
			start := time.Now()
			resp, err := candidate.Provider.Complete(attemptCtx, req.Messages)
			latency := time.Since(start)
			cancel()
			if err == nil {
				c.health.RecordSuccess(candidate.Name, latency)
				if resp.Provider == "" {
					resp.Provider = candidate.Name
				}
				if resp.Model == "" {
					resp.Model = model
				}
				return resp, nil
			}
			c.health.RecordFailure(candidate.Name, latency, err)
			attempts = append(attempts, AttemptError{Provider: candidate.Name, Attempt: attempt + 1, Err: err})
			if ctx.Err() != nil {
				return CompletionResponse{}, CombinedError{Attempts: attempts}
			}
		}
	}
	return CompletionResponse{}, CombinedError{Attempts: attempts}
}

func (c *Client) raceCompletion(ctx context.Context, model string, req ChatRequest, ranked []ProviderInfo) (CompletionResponse, []AttemptError, bool) {
	width := c.options.raceWidth
	if width > len(ranked) {
		width = len(ranked)
	}
	if width < 1 {
		width = 1
	}
	raceCtx, cancelRace := context.WithCancel(ctx)
	defer cancelRace()
	type raceResult struct {
		candidate ProviderInfo
		response  CompletionResponse
		err       error
		latency   time.Duration
	}
	results := make(chan raceResult, width)
	var wg sync.WaitGroup
	for _, candidate := range ranked[:width] {
		if candidate.Provider == nil {
			continue
		}
		wg.Add(1)
		go func(candidate ProviderInfo) {
			defer wg.Done()
			attemptCtx, cancelAttempt := context.WithTimeout(raceCtx, c.options.perAttemptTimeout)
			defer cancelAttempt()
			start := time.Now()
			resp, err := candidate.Provider.Complete(attemptCtx, req.Messages)
			results <- raceResult{candidate: candidate, response: resp, err: err, latency: time.Since(start)}
		}(candidate)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	attempts := make([]AttemptError, 0, width)
	for result := range results {
		if result.err == nil {
			cancelRace()
			wg.Wait()
			c.health.RecordSuccess(result.candidate.Name, result.latency)
			if result.response.Provider == "" {
				result.response.Provider = result.candidate.Name
			}
			if result.response.Model == "" {
				result.response.Model = model
			}
			return result.response, attempts, true
		}
		if raceCtx.Err() != nil && ctx.Err() == nil {
			continue
		}
		c.health.RecordFailure(result.candidate.Name, result.latency, result.err)
		attempts = append(attempts, AttemptError{Provider: result.candidate.Name, Attempt: 1, Err: result.err})
	}
	return CompletionResponse{}, attempts, false
}

// Health returns current provider health snapshots.
func (c *Client) Health() []HealthSnapshot {
	if c == nil || c.health == nil {
		return nil
	}
	return c.health.Snapshot()
}
