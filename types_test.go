package gollmfree

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPublicChatTypesHaveFR4JSONShape(t *testing.T) {
	temperature := 0.7
	maxTokens := 128
	response := CompletionResponse{
		ID:       "cmpl-test",
		Object:   "chat.completion",
		Created:  1710000000,
		Model:    "gpt-3.5-turbo",
		Provider: "fake",
		Choices: []Choice{{
			Index:        0,
			Message:      Message{Role: "assistant", Content: "hello"},
			FinishReason: "stop",
		}},
	}
	chunk := StreamChunk{Content: "hel", Provider: "fake", Model: "gpt-3.5-turbo"}
	req := ChatRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    []Message{{Role: "user", Content: "hi"}},
		Stream:      true,
		Temperature: &temperature,
		MaxTokens:   &maxTokens,
	}

	assertJSONTag(t, Message{}, "Role", "role")
	assertJSONTag(t, Message{}, "Content", "content")
	assertJSONTag(t, ChatRequest{}, "Model", "model")
	assertJSONTag(t, ChatRequest{}, "Messages", "messages")
	assertJSONTag(t, ChatRequest{}, "Stream", "stream,omitempty")
	assertJSONTag(t, ChatRequest{}, "Temperature", "temperature,omitempty")
	assertJSONTag(t, ChatRequest{}, "MaxTokens", "max_tokens,omitempty")
	assertJSONTag(t, CompletionResponse{}, "Provider", "provider,omitempty")
	assertJSONTag(t, Choice{}, "FinishReason", "finish_reason,omitempty")
	assertJSONTag(t, StreamChunk{}, "Content", "content")

	encoded, err := json.Marshal(struct {
		Request  ChatRequest        `json:"request"`
		Response CompletionResponse `json:"response"`
		Chunk    StreamChunk        `json:"chunk"`
	}{Request: req, Response: response, Chunk: chunk})
	if err != nil {
		t.Fatalf("marshal public types: %v", err)
	}

	var decoded struct {
		Request  ChatRequest        `json:"request"`
		Response CompletionResponse `json:"response"`
		Chunk    StreamChunk        `json:"chunk"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal public types: %v", err)
	}
	if decoded.Request.Messages[0].Content != "hi" || decoded.Response.Choices[0].Message.Content != "hello" || decoded.Chunk.Content != "hel" {
		t.Fatalf("round trip lost public chat content: %#v", decoded)
	}
}

func assertJSONTag(t *testing.T, typ any, field, want string) {
	t.Helper()
	got, ok := reflect.TypeOf(typ).FieldByName(field)
	if !ok {
		t.Fatalf("missing field %s on %T", field, typ)
	}
	if tag := got.Tag.Get("json"); tag != want {
		t.Fatalf("%T.%s json tag = %q, want %q", typ, field, tag, want)
	}
}
