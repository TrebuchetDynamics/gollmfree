package gollmfree

import (
	"sort"
	"time"
)

// Selector ranks provider candidates for a requested model.
type Selector struct {
	now func() time.Time
}

// NewSelector constructs a deterministic provider selector.
func NewSelector() *Selector {
	return &Selector{now: time.Now}
}

// Rank returns a ranked copy of candidates for model using static metadata,
// caller priority overrides, and health snapshots.
func (s *Selector) Rank(model string, candidates []ProviderInfo, health []HealthSnapshot, options clientOptions) []ProviderInfo {
	ranked := copyProviderInfos(candidates)
	if len(ranked) < 2 {
		return ranked
	}
	now := time.Now()
	if s != nil && s.now != nil {
		now = s.now()
	}
	model = normalizeRegistryKey(model)
	healthByProvider := make(map[string]HealthSnapshot, len(health))
	for _, snapshot := range health {
		healthByProvider[normalizeRegistryKey(snapshot.Provider)] = snapshot
	}
	priority := priorityIndex(options.providerPriority[model])

	sort.SliceStable(ranked, func(i, j int) bool {
		left, right := ranked[i], ranked[j]
		leftKey, rightKey := normalizeRegistryKey(left.Name), normalizeRegistryKey(right.Name)
		leftOverride, leftHasOverride := priority[leftKey]
		rightOverride, rightHasOverride := priority[rightKey]
		if leftHasOverride != rightHasOverride {
			return leftHasOverride
		}
		if leftHasOverride && leftOverride != rightOverride {
			return leftOverride < rightOverride
		}

		leftHealth, rightHealth := healthByProvider[leftKey], healthByProvider[rightKey]
		leftCooldown := activeCooldown(leftHealth, now)
		rightCooldown := activeCooldown(rightHealth, now)
		if leftCooldown != rightCooldown {
			return !leftCooldown
		}
		if left.DefaultPriority != right.DefaultPriority {
			return left.DefaultPriority < right.DefaultPriority
		}
		if leftHealth.ConsecutiveFailures != rightHealth.ConsecutiveFailures {
			return leftHealth.ConsecutiveFailures < rightHealth.ConsecutiveFailures
		}
		if leftHealth.LastLatency > 0 && rightHealth.LastLatency > 0 && leftHealth.LastLatency != rightHealth.LastLatency {
			return leftHealth.LastLatency < rightHealth.LastLatency
		}
		return leftKey < rightKey
	})
	return ranked
}

func priorityIndex(providers []string) map[string]int {
	out := make(map[string]int, len(providers))
	for index, provider := range providers {
		provider = normalizeRegistryKey(provider)
		if provider == "" {
			continue
		}
		if _, exists := out[provider]; !exists {
			out[provider] = index
		}
	}
	return out
}

func activeCooldown(snapshot HealthSnapshot, now time.Time) bool {
	return !snapshot.CooldownUntil.IsZero() && snapshot.CooldownUntil.After(now)
}
