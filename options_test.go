package gollmfree

import (
	"reflect"
	"testing"
	"time"
)

func TestNewClientAppliesDefaultOptions(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.registry == nil {
		t.Fatal("NewClient did not create a default registry")
	}
	if client.options.defaultModel != "auto" {
		t.Fatalf("default model = %q, want auto", client.options.defaultModel)
	}
	if client.options.perAttemptTimeout != 15*time.Second {
		t.Fatalf("per-attempt timeout = %s, want 15s", client.options.perAttemptTimeout)
	}
	if client.options.maxRetries != 0 {
		t.Fatalf("max retries = %d, want 0", client.options.maxRetries)
	}
	if client.options.raceMode {
		t.Fatal("race mode default = true, want false")
	}
	if client.options.raceWidth != 2 {
		t.Fatalf("race width = %d, want 2", client.options.raceWidth)
	}
}

func TestNewClientAppliesValidOptions(t *testing.T) {
	client := NewClient(
		WithTimeout(3*time.Second),
		WithMaxRetries(2),
		WithRaceMode(true),
		WithRaceWidth(4),
		WithProviderPriority(" GPT-3.5-TURBO ", []string{"DeepAI", " Yqcloud "}),
	)
	if client.options.perAttemptTimeout != 3*time.Second {
		t.Fatalf("per-attempt timeout = %s, want 3s", client.options.perAttemptTimeout)
	}
	if client.options.maxRetries != 2 {
		t.Fatalf("max retries = %d, want 2", client.options.maxRetries)
	}
	if !client.options.raceMode || client.options.raceWidth != 4 {
		t.Fatalf("race options = enabled %v width %d, want true/4", client.options.raceMode, client.options.raceWidth)
	}
	want := []string{"deepai", "yqcloud"}
	if got := client.options.providerPriority["gpt-3.5-turbo"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("provider priority = %#v, want %#v", got, want)
	}
}

func TestNewClientIgnoresInvalidOptions(t *testing.T) {
	client := NewClient(
		WithTimeout(0),
		WithMaxRetries(-1),
		WithRaceWidth(0),
		WithProviderPriority(" ", []string{"DeepAI"}),
		WithProviderPriority("auto", []string{" ", "DeepAI", "deepai"}),
	)
	if client.options.perAttemptTimeout != 15*time.Second {
		t.Fatalf("invalid timeout changed default to %s", client.options.perAttemptTimeout)
	}
	if client.options.maxRetries != 0 {
		t.Fatalf("invalid retry count changed default to %d", client.options.maxRetries)
	}
	if client.options.raceWidth != 2 {
		t.Fatalf("invalid race width changed default to %d", client.options.raceWidth)
	}
	if _, ok := client.options.providerPriority[""]; ok {
		t.Fatal("blank model priority was stored")
	}
	want := []string{"deepai"}
	if got := client.options.providerPriority["auto"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("sanitized auto priority = %#v, want %#v", got, want)
	}
}
