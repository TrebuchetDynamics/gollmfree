package gollmfree

import (
	"strings"
	"time"
)

// Option configures a Client during construction.
type Option func(*clientOptions)

type clientOptions struct {
	defaultModel      string
	perAttemptTimeout time.Duration
	maxRetries        int
	raceMode          bool
	raceWidth         int
	providerPriority  map[string][]string
	registryOverride  *Registry
}

func defaultClientOptions() clientOptions {
	return clientOptions{
		defaultModel:      "auto",
		perAttemptTimeout: 15 * time.Second,
		maxRetries:        0,
		raceMode:          false,
		raceWidth:         2,
		providerPriority:  make(map[string][]string),
	}
}

// WithTimeout sets the per-provider attempt timeout. Non-positive durations are
// ignored so invalid configuration cannot remove the default timeout bound.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		if timeout <= 0 {
			return
		}
		options.perAttemptTimeout = timeout
	}
}

// WithMaxRetries sets the number of retries for each provider after the initial
// attempt. Negative values are ignored.
func WithMaxRetries(retries int) Option {
	return func(options *clientOptions) {
		if retries < 0 {
			return
		}
		options.maxRetries = retries
	}
}

// WithRaceMode enables or disables race mode. Race mode is disabled by default
// to avoid extra traffic to anonymous providers.
func WithRaceMode(enabled bool) Option {
	return func(options *clientOptions) {
		options.raceMode = enabled
	}
}

// WithRaceWidth sets how many providers may be started concurrently when race
// mode is enabled. Values less than one are ignored.
func WithRaceWidth(width int) Option {
	return func(options *clientOptions) {
		if width < 1 {
			return
		}
		options.raceWidth = width
	}
}

// WithRegistry replaces the default (empty) registry with a pre-built one.
// This is the primary way to wire concrete providers into NewClient from an
// external package, since the registry field is unexported.
func WithRegistry(r *Registry) Option {
	return func(opts *clientOptions) {
		opts.registryOverride = r
	}
}

// WithProviderPriority sets a preferred provider order for a model alias.
// Model and provider names are normalized by trimming whitespace and lowercasing;
// blank and duplicate provider names are ignored.
func WithProviderPriority(model string, providers []string) Option {
	return func(options *clientOptions) {
		model = normalizeRegistryKey(model)
		if model == "" {
			return
		}
		cleaned := normalizePriorityProviders(providers)
		if len(cleaned) == 0 {
			delete(options.providerPriority, model)
			return
		}
		options.providerPriority[model] = cleaned
	}
}

func normalizePriorityProviders(providers []string) []string {
	seen := make(map[string]struct{}, len(providers))
	cleaned := make([]string, 0, len(providers))
	for _, provider := range providers {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" {
			continue
		}
		if _, exists := seen[provider]; exists {
			continue
		}
		seen[provider] = struct{}{}
		cleaned = append(cleaned, provider)
	}
	return cleaned
}
