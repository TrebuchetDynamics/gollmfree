package gollmfree

import (
	"fmt"
	"sort"
	"strings"
)

// Registry is an immutable collection of providers and model aliases.
// It is safe for concurrent reads after construction.
type Registry struct {
	providers  []ProviderInfo
	byName     map[string]ProviderInfo
	byModel    map[string][]ProviderInfo
	aliasOrder []string
}

// NewRegistry creates an immutable provider registry. Provider names and model
// aliases are normalized for lookup by trimming whitespace and lowercasing.
func NewRegistry(infos ...ProviderInfo) (*Registry, error) {
	registry := &Registry{
		providers: make([]ProviderInfo, 0, len(infos)),
		byName:    make(map[string]ProviderInfo, len(infos)),
		byModel:   make(map[string][]ProviderInfo),
	}
	seenAlias := make(map[string]struct{})

	for _, info := range infos {
		stored := copyProviderInfo(info)
		stored.Name = strings.TrimSpace(stored.Name)
		if stored.Name == "" {
			return nil, fmt.Errorf("gollmfree: provider name is required")
		}
		key := normalizeRegistryKey(stored.Name)
		if _, exists := registry.byName[key]; exists {
			return nil, fmt.Errorf("gollmfree: duplicate provider name %q", stored.Name)
		}

		stored.SupportedModels = normalizeModelAliases(stored.SupportedModels)
		registry.providers = append(registry.providers, stored)
		registry.byName[key] = stored

		aliases := append([]string{key, "auto", "best"}, stored.SupportedModels...)
		seenProviderAlias := make(map[string]struct{}, len(aliases))
		for _, alias := range aliases {
			alias = normalizeRegistryKey(alias)
			if alias == "" {
				continue
			}
			if _, exists := seenProviderAlias[alias]; exists {
				continue
			}
			seenProviderAlias[alias] = struct{}{}
			if _, exists := seenAlias[alias]; !exists {
				seenAlias[alias] = struct{}{}
				registry.aliasOrder = append(registry.aliasOrder, alias)
			}
			registry.byModel[alias] = append(registry.byModel[alias], stored)
		}
	}

	return registry, nil
}

// Providers returns registered providers in registration order.
func (r *Registry) Providers() []ProviderInfo {
	if r == nil {
		return nil
	}
	return copyProviderInfos(r.providers)
}

// Provider returns provider metadata by name using normalized lookup.
func (r *Registry) Provider(name string) (ProviderInfo, bool) {
	if r == nil {
		return ProviderInfo{}, false
	}
	info, ok := r.byName[normalizeRegistryKey(name)]
	if !ok {
		return ProviderInfo{}, false
	}
	return copyProviderInfo(info), true
}

// Candidates returns providers that can handle the requested model alias.
func (r *Registry) Candidates(model string) []ProviderInfo {
	if r == nil {
		return nil
	}
	return copyProviderInfos(r.byModel[normalizeRegistryKey(model)])
}

// Models returns known aliases and provider coverage in deterministic alias order.
func (r *Registry) Models() []ModelInfo {
	if r == nil {
		return nil
	}
	models := make([]ModelInfo, 0, len(r.aliasOrder))
	for _, alias := range r.aliasOrder {
		providers := r.byModel[alias]
		if len(providers) == 0 {
			continue
		}
		model := ModelInfo{Alias: alias, Providers: make([]string, len(providers))}
		for i, provider := range providers {
			model.Providers[i] = provider.Name
		}
		models = append(models, model)
	}
	return models
}

func normalizeRegistryKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeModelAliases(aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(aliases))
	normalized := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		alias = normalizeRegistryKey(alias)
		if alias == "" {
			continue
		}
		if _, exists := seen[alias]; exists {
			continue
		}
		seen[alias] = struct{}{}
		normalized = append(normalized, alias)
	}
	sort.Strings(normalized)
	return normalized
}

func copyProviderInfos(infos []ProviderInfo) []ProviderInfo {
	if len(infos) == 0 {
		return nil
	}
	copies := make([]ProviderInfo, len(infos))
	for i, info := range infos {
		copies[i] = copyProviderInfo(info)
	}
	return copies
}

func copyProviderInfo(info ProviderInfo) ProviderInfo {
	copied := info
	if info.SupportedModels != nil {
		copied.SupportedModels = append([]string(nil), info.SupportedModels...)
	}
	return copied
}
