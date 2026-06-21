package main

import (
	"strings"
	"testing"
)

func TestRunNoArgsReturnsExitCode2(t *testing.T) {
	code := run(nil)
	if code != 2 {
		t.Fatalf("run() with no args = %d, want 2", code)
	}
}

func TestRunUnknownCommandReturnsExitCode2(t *testing.T) {
	code := run([]string{"boguscommand"})
	if code != 2 {
		t.Fatalf("run(boguscommand) = %d, want 2", code)
	}
}

func TestCmdChatNoPromptReturnsExitCode2(t *testing.T) {
	code := cmdChat(nil)
	if code != 2 {
		t.Fatalf("cmdChat with no prompt = %d, want 2", code)
	}
}

func TestCmdListPrintsProvidersWithoutNetwork(t *testing.T) {
	// list requires no network — it reads from the registry only.
	code := cmdList(nil)
	if code != 0 {
		t.Fatalf("cmdList = %d, want 0", code)
	}
}

func TestCmdModelsPrintsAliasesWithoutNetwork(t *testing.T) {
	code := cmdModels(nil)
	if code != 0 {
		t.Fatalf("cmdModels = %d, want 0", code)
	}
}

func TestDefaultRegistryContainsExpectedProviders(t *testing.T) {
	reg := defaultRegistry()
	providerNames := make([]string, 0)
	for _, p := range reg.Providers() {
		providerNames = append(providerNames, p.Name)
	}
	want := []string{"pollinationsai", "chatai", "yqcloud", "wewordle"}
	for _, name := range want {
		found := false
		for _, got := range providerNames {
			if got == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("defaultRegistry missing provider %q; got %s", name, strings.Join(providerNames, ","))
		}
	}
}

func TestDefaultRegistryModelsIncludeAutoAlias(t *testing.T) {
	reg := defaultRegistry()
	for _, m := range reg.Models() {
		if m.Alias == "auto" {
			return
		}
	}
	t.Fatal("defaultRegistry models missing 'auto' alias")
}
