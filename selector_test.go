package gollmfree

import (
	"reflect"
	"testing"
	"time"
)

func TestSelectorRankOrdersByOverrideCooldownPriorityFailuresLatencyName(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	selector := NewSelector()
	selector.now = func() time.Time { return now }

	candidates := []ProviderInfo{
		{Name: "zeta", DefaultPriority: 20},
		{Name: "alpha", DefaultPriority: 10},
		{Name: "slow", DefaultPriority: 10},
		{Name: "failed", DefaultPriority: 10},
		{Name: "cooldown", DefaultPriority: 1},
		{Name: "override", DefaultPriority: 99},
	}
	health := []HealthSnapshot{
		{Provider: "slow", LastLatency: 250 * time.Millisecond},
		{Provider: "failed", ConsecutiveFailures: 2, LastLatency: 10 * time.Millisecond},
		{Provider: "cooldown", CooldownUntil: now.Add(time.Minute)},
	}
	options := defaultClientOptions()
	options.providerPriority["auto"] = []string{"override"}

	ranked := selector.Rank("auto", candidates, health, options)
	got := providerNames(ranked)
	want := []string{"override", "alpha", "slow", "failed", "zeta", "cooldown"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked providers = %#v, want %#v", got, want)
	}
}

func TestSelectorRankKeepsExpiredCooldownInNormalOrder(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	selector := NewSelector()
	selector.now = func() time.Time { return now }

	candidates := []ProviderInfo{
		{Name: "healthy", DefaultPriority: 10},
		{Name: "recovered", DefaultPriority: 1},
	}
	health := []HealthSnapshot{{Provider: "recovered", CooldownUntil: now.Add(-time.Second)}}

	ranked := selector.Rank("auto", candidates, health, defaultClientOptions())
	got := providerNames(ranked)
	want := []string{"recovered", "healthy"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked providers = %#v, want %#v", got, want)
	}
}
