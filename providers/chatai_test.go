package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

func TestChataiCompletePostsUpstreamShapeAndParsesSSEDeltas(t *testing.T) {
	var gotBody chataiRequest
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
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	provider := NewChatai(WithChataiEndpoint(server.URL), WithChataiToken("test-token"))
	resp, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if gotBody.Token != "test-token" {
		t.Fatalf("request token = %q, want test-token", gotBody.Token)
	}
	if gotBody.Type != 1 {
		t.Fatalf("request type = %d, want 1", gotBody.Type)
	}
	if gotBody.MachineID == "" {
		t.Fatal("request machineId is empty")
	}
	if len(gotBody.Msg) != 1 || gotBody.Msg[0].Content != "hi" {
		t.Fatalf("request msg = %#v", gotBody.Msg)
	}
	if resp.Provider != chataiName {
		t.Fatalf("response provider = %q, want %q", resp.Provider, chataiName)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "hello world" {
		t.Fatalf("response content = %q, want \"hello world\"", func() string {
			if len(resp.Choices) > 0 {
				return resp.Choices[0].Message.Content
			}
			return ""
		}())
	}
}

func TestChataiCompleteSkipsNonDataSSELines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": comment line\n\n")
		fmt.Fprint(w, "event: message\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	provider := NewChatai(WithChataiEndpoint(server.URL))
	resp, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "ok" {
		t.Fatalf("response content = %q, want \"ok\"", func() string {
			if len(resp.Choices) > 0 {
				return resp.Choices[0].Message.Content
			}
			return ""
		}())
	}
}

func TestChataiCompleteReturnsProviderNamedErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	provider := NewChatai(WithChataiEndpoint(server.URL))
	_, err := provider.Complete(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for non-2xx response")
	}
	if got := err.Error(); !strings.Contains(got, chataiName) || !strings.Contains(got, "429") {
		t.Fatalf("non-2xx error = %q, want provider name and status", got)
	}
}

func TestChataiCompletePreservesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewChatai(WithChataiEndpoint(server.URL))
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()
	_, err := provider.Complete(ctx, []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("Complete returned nil error for canceled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v, want context cancellation", err)
	}
	if !strings.Contains(err.Error(), chataiName) {
		t.Fatalf("error = %q, want provider name", err.Error())
	}
}

func TestChataiStreamEmitsAccumulatedSSEContentAndCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"chunk1\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" chunk2\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	provider := NewChatai(WithChataiEndpoint(server.URL))
	stream, err := provider.Stream(t.Context(), []gollmfree.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	chunk, ok := <-stream
	if !ok || chunk != "chunk1 chunk2" {
		t.Fatalf("stream chunk = %q, %v; want \"chunk1 chunk2\", true", chunk, ok)
	}
	if _, ok := <-stream; ok {
		t.Fatal("stream emitted extra chunks after first")
	}
}
