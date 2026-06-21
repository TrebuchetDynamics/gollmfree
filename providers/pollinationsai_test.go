package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

func TestPollinationsAICompletePostsOpenAIShapeAndParsesResponse(t *testing.T) {
	var gotRequest struct {
		Model    string              `json:"model"`
		Messages []gollmfree.Message `json:"messages"`
		Stream   bool                `json:"stream"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"chatcmpl-test",
			"object":"chat.completion",
			"created":1710000000,
			"model":"openai-fast",
			"choices":[{"index":0,"message":{"role":"assistant","content":"hello from pollinations"},"finish_reason":"stop"}]
		}`))
	}))
	defer server.Close()

	provider := NewPollinationsAI(WithPollinationsEndpoint(server.URL))
	resp, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if gotRequest.Model != "openai-fast" {
		t.Fatalf("request model = %q, want openai-fast", gotRequest.Model)
	}
	if gotRequest.Stream {
		t.Fatal("request stream = true, want false for Complete")
	}
	if len(gotRequest.Messages) != 1 || gotRequest.Messages[0].Content != "hello" {
		t.Fatalf("request messages = %#v", gotRequest.Messages)
	}
	if resp.Provider != "pollinationsai" {
		t.Fatalf("response provider = %q, want pollinationsai", resp.Provider)
	}
	if resp.Model != "openai-fast" || resp.Choices[0].Message.Content != "hello from pollinations" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestPollinationsAICompleteReturnsProviderNamedErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream quota exceeded with a long diagnostic body", http.StatusTooManyRequests)
	}))
	defer server.Close()

	provider := NewPollinationsAI(WithPollinationsEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("Complete returned nil error for non-2xx response")
	}
	if got := err.Error(); !strings.Contains(got, "pollinationsai") || !strings.Contains(got, "429") || !strings.Contains(got, "upstream quota exceeded") {
		t.Fatalf("non-2xx error = %q, want provider name, status, and body snippet", got)
	}
}

func TestPollinationsAICompleteReturnsProviderNamedErrorForMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[`))
	}))
	defer server.Close()

	provider := NewPollinationsAI(WithPollinationsEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("Complete returned nil error for malformed JSON")
	}
	if got := err.Error(); !strings.Contains(got, "pollinationsai") || !strings.Contains(got, "decode response") {
		t.Fatalf("malformed JSON error = %q, want provider-named decode error", got)
	}
}

func TestPollinationsAICompletePreservesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewPollinationsAI(WithPollinationsEndpoint(server.URL))
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()
	_, err := provider.Complete(ctx, []gollmfree.Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("Complete returned nil error for canceled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v, want context cancellation detectable", err)
	}
	if got := err.Error(); !strings.Contains(got, "pollinationsai") {
		t.Fatalf("cancellation error = %q, want provider name", got)
	}
}

func TestPollinationsAIStreamEmitsOneCompleteChunkAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gotRequest struct {
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if gotRequest.Stream {
			t.Fatal("Stream emulation should call non-streaming completion with stream=false")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model":"openai-fast",
			"choices":[{"index":0,"message":{"role":"assistant","content":"complete stream chunk"},"finish_reason":"stop"}]
		}`))
	}))
	defer server.Close()

	provider := NewPollinationsAI(WithPollinationsEndpoint(server.URL))
	stream, err := provider.Stream(t.Context(), []gollmfree.Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	chunk, ok := <-stream
	if !ok || chunk != "complete stream chunk" {
		t.Fatalf("first stream receive = %q, %v; want complete stream chunk, true", chunk, ok)
	}
	if chunk, ok := <-stream; ok {
		t.Fatalf("stream emitted extra chunk %q", chunk)
	}
}
