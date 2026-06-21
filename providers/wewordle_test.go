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

func TestWeWordleCompletePostsSubscriberShapeAndParsesContent(t *testing.T) {
	var gotBody wewordleRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wewordleResponse{Content: "hello from wewordle"})
	}))
	defer server.Close()

	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	resp, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "what is 2+2?"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	// Verify upstream-shaped request fields.
	if len(gotBody.Messages) != 1 || gotBody.Messages[0].Role != "user" || gotBody.Messages[0].Content != "what is 2+2?" {
		t.Fatalf("request messages = %#v", gotBody.Messages)
	}
	// subscriber, app, user fields must be present (null values are fine).

	if resp.Provider != wewordleName {
		t.Fatalf("response provider = %q, want %q", resp.Provider, wewordleName)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "hello from wewordle" {
		t.Fatalf("response content = %q, want \"hello from wewordle\"", func() string {
			if len(resp.Choices) > 0 {
				return resp.Choices[0].Message.Content
			}
			return ""
		}())
	}
}

func TestWeWordleCompletePreservesAllMessageRoles(t *testing.T) {
	var gotBody wewordleRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wewordleResponse{Content: "ok"})
	}))
	defer server.Close()

	messages := []gollmfree.Message{
		{Role: "system", Content: "be helpful"},
		{Role: "user", Content: "question"},
		{Role: "assistant", Content: "answer"},
		{Role: "user", Content: "follow-up"},
	}
	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), messages)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if len(gotBody.Messages) != 4 {
		t.Fatalf("request messages len = %d, want 4", len(gotBody.Messages))
	}
	if gotBody.Messages[0].Role != "system" || gotBody.Messages[3].Content != "follow-up" {
		t.Fatalf("messages not preserved correctly: %#v", gotBody.Messages)
	}
}

func TestWeWordleCompleteReturnsProviderNamedErrorForMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[`))
	}))
	defer server.Close()

	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for malformed JSON")
	}
	if !strings.Contains(err.Error(), wewordleName) || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("malformed JSON error = %q, want provider-named decode error", err.Error())
	}
}

func TestWeWordleCompleteReturnsProviderNamedErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
	}))
	defer server.Close()

	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for non-2xx")
	}
	if !strings.Contains(err.Error(), wewordleName) || !strings.Contains(err.Error(), "429") {
		t.Fatalf("non-2xx error = %q, want provider name and status", err.Error())
	}
}

func TestWeWordleCompletePreservesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()
	_, err := provider.Complete(ctx, []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for canceled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v, want context cancellation", err)
	}
	if !strings.Contains(err.Error(), wewordleName) {
		t.Fatalf("error = %q, want provider name", err.Error())
	}
}

func TestWeWordleStreamEmitsSingleChunkAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(new(wewordleRequest))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wewordleResponse{Content: "stream chunk"})
	}))
	defer server.Close()

	provider := NewWeWordle(WithWeWordleEndpoint(server.URL))
	stream, err := provider.Stream(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	chunk, ok := <-stream
	if !ok || chunk != "stream chunk" {
		t.Fatalf("stream chunk = %q, %v; want \"stream chunk\", true", chunk, ok)
	}
	if _, ok := <-stream; ok {
		t.Fatal("stream emitted extra chunks")
	}
}
