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

func TestYqcloudCompletePostsPromptAndParsesPlainTextStream(t *testing.T) {
	var gotBody yqcloudRequest
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
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello from yqcloud"))
	}))
	defer server.Close()

	provider := NewYqcloud(WithYqcloudEndpoint(server.URL))
	resp, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "what is 2+2?"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if gotBody.Prompt != "what is 2+2?" {
		t.Fatalf("request prompt = %q, want last user message content", gotBody.Prompt)
	}
	if gotBody.UserID == "" {
		t.Fatal("request userId is empty")
	}
	if resp.Provider != yqcloudName {
		t.Fatalf("response provider = %q, want %q", resp.Provider, yqcloudName)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "hello from yqcloud" {
		t.Fatalf("response content = %q, want \"hello from yqcloud\"", func() string {
			if len(resp.Choices) > 0 {
				return resp.Choices[0].Message.Content
			}
			return ""
		}())
	}
}

func TestYqcloudCompleteUsesLastUserMessageAsPrompt(t *testing.T) {
	var gotBody yqcloudRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	messages := []gollmfree.Message{
		{Role: "system", Content: "be helpful"},
		{Role: "user", Content: "first question"},
		{Role: "assistant", Content: "first answer"},
		{Role: "user", Content: "second question"},
	}
	provider := NewYqcloud(WithYqcloudEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), messages)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if gotBody.Prompt != "second question" {
		t.Fatalf("prompt = %q, want last user message content", gotBody.Prompt)
	}
}

func TestYqcloudCompleteReturnsProviderNamedErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := NewYqcloud(WithYqcloudEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for non-2xx")
	}
	if got := err.Error(); !strings.Contains(got, yqcloudName) || !strings.Contains(got, "503") {
		t.Fatalf("non-2xx error = %q, want provider name and status", got)
	}
}

func TestYqcloudCompletePreservesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewYqcloud(WithYqcloudEndpoint(server.URL))
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()
	_, err := provider.Complete(ctx, []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for canceled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v, want context cancellation", err)
	}
	if !strings.Contains(err.Error(), yqcloudName) {
		t.Fatalf("error = %q, want provider name", err.Error())
	}
}

func TestYqcloudStreamEmitsSingleChunkAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(new(yqcloudRequest))
		_, _ = w.Write([]byte("streamed response"))
	}))
	defer server.Close()

	provider := NewYqcloud(WithYqcloudEndpoint(server.URL))
	stream, err := provider.Stream(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	chunk, ok := <-stream
	if !ok || chunk != "streamed response" {
		t.Fatalf("stream chunk = %q, %v; want \"streamed response\", true", chunk, ok)
	}
	if _, ok := <-stream; ok {
		t.Fatal("stream emitted extra chunks")
	}
}
