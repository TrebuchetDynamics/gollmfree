package gollmfree

import (
	"context"
	"reflect"
	"testing"
)

func TestRegistryLookupAliasesAndUnknownModel(t *testing.T) {
	registry, err := NewRegistry(
		ProviderInfo{Name: "DeepAI", Provider: registryProvider{name: "DeepAI"}, SupportedModels: []string{"Auto", "GPT-3.5-Turbo"}, DefaultPriority: 1},
		ProviderInfo{Name: "Yqcloud", Provider: registryProvider{name: "Yqcloud"}, SupportedModels: []string{"best", "gpt-3.5-turbo"}, DefaultPriority: 3},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	provider, ok := registry.Provider(" deepai ")
	if !ok || provider.Name != "DeepAI" {
		t.Fatalf("Provider lookup = %#v, %v; want DeepAI, true", provider, ok)
	}

	got := providerNames(registry.Candidates(" GPT-3.5-TURBO "))
	want := []string{"DeepAI", "Yqcloud"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates(gpt-3.5-turbo) = %#v, want %#v", got, want)
	}

	got = providerNames(registry.Candidates("deepai"))
	want = []string{"DeepAI"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates(deepai) = %#v, want %#v", got, want)
	}

	got = providerNames(registry.Candidates("best"))
	want = []string{"DeepAI", "Yqcloud"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates(best) = %#v, want %#v", got, want)
	}

	if got := registry.Candidates("unknown-model"); len(got) != 0 {
		t.Fatalf("Candidates(unknown-model) = %#v, want empty", got)
	}
}

func TestRegistryRejectsDuplicateProviderNames(t *testing.T) {
	_, err := NewRegistry(
		ProviderInfo{Name: "DeepAI", Provider: registryProvider{name: "DeepAI"}},
		ProviderInfo{Name: " deepai ", Provider: registryProvider{name: "deepai"}},
	)
	if err == nil {
		t.Fatal("NewRegistry accepted duplicate normalized provider names")
	}
}

func TestRegistryReturnsDefensiveCopies(t *testing.T) {
	registry, err := NewRegistry(ProviderInfo{Name: "DeepAI", Provider: registryProvider{name: "DeepAI"}, SupportedModels: []string{"auto", "gpt-3.5-turbo"}, DefaultPriority: 1})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	providers := registry.Providers()
	providers[0].Name = "mutated"
	providers[0].SupportedModels[0] = "mutated"
	if got := registry.Providers()[0]; got.Name != "DeepAI" || got.SupportedModels[0] != "auto" {
		t.Fatalf("Providers returned mutable registry state: %#v", got)
	}

	models := registry.Models()
	models[0].Alias = "mutated"
	models[0].Providers[0] = "mutated"
	if got := registry.Models()[0]; got.Alias == "mutated" || got.Providers[0] != "DeepAI" {
		t.Fatalf("Models returned mutable registry state: %#v", got)
	}

	candidates := registry.Candidates("auto")
	candidates[0].Name = "mutated"
	candidates[0].SupportedModels[0] = "mutated"
	if got := registry.Candidates("auto")[0]; got.Name != "DeepAI" || got.SupportedModels[0] != "auto" {
		t.Fatalf("Candidates returned mutable registry state: %#v", got)
	}
}

func providerNames(infos []ProviderInfo) []string {
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name
	}
	return names
}

type registryProvider struct{ name string }

func (p registryProvider) Name() string { return p.name }

func (p registryProvider) Complete(ctx context.Context, messages []Message) (CompletionResponse, error) {
	return CompletionResponse{}, nil
}

func (p registryProvider) Stream(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (p registryProvider) SupportedModels() []string { return nil }
