package gollmfree

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientChatCompletionFallsBackAfterProviderFailure(t *testing.T) {
	firstErr := errors.New("first failed")
	first := &clientFakeProvider{name: "first", err: firstErr}
	second := &clientFakeProvider{name: "second", response: "ok from second"}
	registry, err := NewRegistry(
		ProviderInfo{Name: "first", Provider: first, SupportedModels: []string{"auto"}, DefaultPriority: 1},
		ProviderInfo{Name: "second", Provider: second, SupportedModels: []string{"auto"}, DefaultPriority: 2},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	client := NewClient()
	client.registry = registry
	client.health = NewHealthStore(3, time.Minute)

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatalf("ChatCompletion returned error: %v", err)
	}
	if resp.Provider != "second" || resp.Choices[0].Message.Content != "ok from second" {
		t.Fatalf("response = %#v", resp)
	}
	if first.CompleteCalls() != 1 || second.CompleteCalls() != 1 {
		t.Fatalf("calls = first %d second %d, want 1/1", first.CompleteCalls(), second.CompleteCalls())
	}
	snaps := client.Health()
	if len(snaps) != 2 || snaps[0].Provider != "first" || snaps[0].Failures != 1 || snaps[1].Provider != "second" || snaps[1].Successes != 1 {
		t.Fatalf("health snapshots = %#v", snaps)
	}
}

func TestClientChatCompletionReturnsCombinedErrorWhenAllProvidersFail(t *testing.T) {
	first := &clientFakeProvider{name: "first", err: errors.New("first failed")}
	second := &clientFakeProvider{name: "second", err: errors.New("second failed")}
	registry, err := NewRegistry(
		ProviderInfo{Name: "first", Provider: first, SupportedModels: []string{"auto"}, DefaultPriority: 1},
		ProviderInfo{Name: "second", Provider: second, SupportedModels: []string{"auto"}, DefaultPriority: 2},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	client := NewClient()
	client.registry = registry
	client.health = NewHealthStore(3, time.Minute)

	_, err = client.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}})
	if err == nil {
		t.Fatal("ChatCompletion returned nil error when every provider failed")
	}
	got := err.Error()
	for _, want := range []string{"first", "first failed", "second", "second failed"} {
		if !strings.Contains(got, want) {
			t.Fatalf("combined error = %q, missing %q", got, want)
		}
	}
}

func TestClientChatCompletionRetriesProviderBeforeFallback(t *testing.T) {
	flaky := &clientFakeProvider{name: "flaky", response: "ok after retry", err: errors.New("transient"), failFor: 1}
	fallback := &clientFakeProvider{name: "fallback", response: "fallback ok"}
	registry, err := NewRegistry(
		ProviderInfo{Name: "flaky", Provider: flaky, SupportedModels: []string{"auto"}, DefaultPriority: 1},
		ProviderInfo{Name: "fallback", Provider: fallback, SupportedModels: []string{"auto"}, DefaultPriority: 2},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	client := NewClient(WithMaxRetries(1))
	client.registry = registry
	client.health = NewHealthStore(3, time.Minute)

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatalf("ChatCompletion returned error: %v", err)
	}
	if resp.Provider != "flaky" || resp.Choices[0].Message.Content != "ok after retry" {
		t.Fatalf("response = %#v", resp)
	}
	if flaky.CompleteCalls() != 2 || fallback.CompleteCalls() != 0 {
		t.Fatalf("calls = flaky %d fallback %d, want 2/0", flaky.CompleteCalls(), fallback.CompleteCalls())
	}
}

func TestClientChatCompletionPerAttemptTimeoutFallsBack(t *testing.T) {
	slow := &clientFakeProvider{name: "slow", response: "too late", delay: time.Minute}
	fast := &clientFakeProvider{name: "fast", response: "fast ok"}
	registry, err := NewRegistry(
		ProviderInfo{Name: "slow", Provider: slow, SupportedModels: []string{"auto"}, DefaultPriority: 1},
		ProviderInfo{Name: "fast", Provider: fast, SupportedModels: []string{"auto"}, DefaultPriority: 2},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	client := NewClient(WithTimeout(time.Millisecond))
	client.registry = registry
	client.health = NewHealthStore(3, time.Minute)

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatalf("ChatCompletion returned error: %v", err)
	}
	if resp.Provider != "fast" || slow.CompleteCalls() != 1 || fast.CompleteCalls() != 1 {
		t.Fatalf("response/calls = %#v slow %d fast %d, want fast and 1/1", resp, slow.CompleteCalls(), fast.CompleteCalls())
	}
	if snaps := client.Health(); len(snaps) != 2 || snaps[1].Provider != "slow" || !strings.Contains(snaps[1].LastError, "deadline") {
		t.Fatalf("health snapshots = %#v, want slow deadline failure", snaps)
	}
}

func TestClientChatCompletionRaceModeFirstSuccessWinsWithoutCanceledLoserFailure(t *testing.T) {
	slow := &clientFakeProvider{name: "slow", response: "slow ok", delay: 50 * time.Millisecond}
	fast := &clientFakeProvider{name: "fast", response: "fast ok"}
	registry, err := NewRegistry(
		ProviderInfo{Name: "slow", Provider: slow, SupportedModels: []string{"auto"}, DefaultPriority: 1},
		ProviderInfo{Name: "fast", Provider: fast, SupportedModels: []string{"auto"}, DefaultPriority: 2},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	client := NewClient(WithRaceMode(true), WithRaceWidth(2))
	client.registry = registry
	client.health = NewHealthStore(3, time.Minute)

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatalf("ChatCompletion returned error: %v", err)
	}
	if resp.Provider != "fast" || slow.CompleteCalls() != 1 || fast.CompleteCalls() != 1 {
		t.Fatalf("response/calls = %#v slow %d fast %d, want fast and both called", resp, slow.CompleteCalls(), fast.CompleteCalls())
	}
	snaps := client.Health()
	if len(snaps) != 1 || snaps[0].Provider != "fast" || snaps[0].Successes != 1 || snaps[0].Failures != 0 {
		t.Fatalf("health snapshots = %#v, want only fast success and no canceled loser failure", snaps)
	}
}

type clientFakeProvider struct {
	name     string
	response string
	err      error
	failFor  int64
	delay    time.Duration
	calls    atomic.Int64
}

func (p *clientFakeProvider) Name() string { return p.name }

func (p *clientFakeProvider) SupportedModels() []string { return []string{"auto"} }

func (p *clientFakeProvider) Complete(ctx context.Context, _ []Message) (CompletionResponse, error) {
	call := p.calls.Add(1)
	if p.delay > 0 {
		timer := time.NewTimer(p.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return CompletionResponse{}, ctx.Err()
		case <-timer.C:
		}
	}
	if p.err != nil && (p.failFor == 0 || call <= p.failFor) {
		return CompletionResponse{}, p.err
	}
	return CompletionResponse{Provider: p.name, Choices: []Choice{{Message: Message{Role: "assistant", Content: p.response}}}}, nil
}

func (p *clientFakeProvider) Stream(context.Context, []Message) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (p *clientFakeProvider) CompleteCalls() int64 { return p.calls.Load() }
