package gollmfree

import (
	"context"
	"testing"
)

var _ Provider = fakeProvider{}

func TestProviderInterfaceAcceptsFakeProvider(t *testing.T) {
	provider := fakeProvider{}
	if provider.Name() != "fake" {
		t.Fatalf("Name() = %q, want fake", provider.Name())
	}
	models := provider.SupportedModels()
	if len(models) != 2 || models[0] != "auto" || models[1] != "fake-model" {
		t.Fatalf("SupportedModels() = %#v", models)
	}
	resp, err := provider.Complete(context.Background(), []Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if resp.Provider != "fake" || resp.Choices[0].Message.Content != "ok" {
		t.Fatalf("Complete response = %#v", resp)
	}
	stream, err := provider.Stream(context.Background(), nil)
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	chunk, ok := <-stream
	if !ok || chunk != "ok" {
		t.Fatalf("first stream receive = %q, %v; want ok, true", chunk, ok)
	}
	if _, ok := <-stream; ok {
		t.Fatal("stream channel remained open after fake provider sent its chunk")
	}
}

func TestProviderMetadataCopies(t *testing.T) {
	info := ProviderInfo{
		Name:            "fake",
		Provider:        fakeProvider{},
		SupportedModels: []string{"auto", "fake-model"},
		DefaultPriority: 10,
	}
	model := ModelInfo{Alias: "auto", Providers: []string{"fake"}}
	if info.Name != "fake" || info.DefaultPriority != 10 || len(info.SupportedModels) != 2 || info.Provider.Name() != "fake" {
		t.Fatalf("ProviderInfo fields not usable: %#v", info)
	}
	if model.Alias != "auto" || len(model.Providers) != 1 || model.Providers[0] != "fake" {
		t.Fatalf("ModelInfo fields not usable: %#v", model)
	}
}

type fakeProvider struct{}

func (fakeProvider) Name() string { return "fake" }

func (fakeProvider) Complete(context.Context, []Message) (CompletionResponse, error) {
	return CompletionResponse{
		Provider: "fake",
		Choices:  []Choice{{Message: Message{Role: "assistant", Content: "ok"}}},
	}, nil
}

func (fakeProvider) Stream(context.Context, []Message) (<-chan string, error) {
	chunks := make(chan string, 1)
	chunks <- "ok"
	close(chunks)
	return chunks, nil
}

func (fakeProvider) SupportedModels() []string { return []string{"auto", "fake-model"} }
