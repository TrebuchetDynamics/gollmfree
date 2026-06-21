package testprovider

import (
	"context"
	"errors"
	"testing"
	"time"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

var _ gollmfree.Provider = New("fake")

func TestFakeProviderCanSimulateSuccessFailureDelayAndStream(t *testing.T) {
	provider := New("fake",
		WithModels("auto", "fake-model"),
		WithResponse("hello"),
		WithStreamChunks("hel", "lo"),
		WithDelay(10*time.Millisecond),
	)

	start := time.Now()
	resp, err := provider.Complete(context.Background(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Fatal("Complete did not apply configured delay")
	}
	if resp.Provider != "fake" || resp.Choices[0].Message.Content != "hello" {
		t.Fatalf("Complete response = %#v", resp)
	}
	if got := provider.CompleteCalls(); got != 1 {
		t.Fatalf("CompleteCalls = %d, want 1", got)
	}

	stream, err := provider.Stream(context.Background(), nil)
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	var chunks []string
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}
	if len(chunks) != 2 || chunks[0] != "hel" || chunks[1] != "lo" {
		t.Fatalf("stream chunks = %#v, want hel/lo", chunks)
	}
	if got := provider.StreamCalls(); got != 1 {
		t.Fatalf("StreamCalls = %d, want 1", got)
	}

	boom := errors.New("boom")
	provider = New("broken", WithError(boom))
	if _, err := provider.Complete(context.Background(), nil); !errors.Is(err, boom) {
		t.Fatalf("Complete error = %v, want boom", err)
	}
	if _, err := provider.Stream(context.Background(), nil); !errors.Is(err, boom) {
		t.Fatalf("Stream error = %v, want boom", err)
	}
}

func TestFakeProviderRespectsContextCancellationDuringDelay(t *testing.T) {
	provider := New("slow", WithDelay(time.Minute), WithResponse("late"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Complete(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Complete cancellation error = %v, want context.Canceled", err)
	}
	if _, err := provider.Stream(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Stream cancellation error = %v, want context.Canceled", err)
	}
}
