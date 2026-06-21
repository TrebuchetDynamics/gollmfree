package gollmfree

import (
	"sort"
	"sync"
	"time"
)

// HealthSnapshot is a point-in-time copy of provider health state.
type HealthSnapshot struct {
	Provider            string
	Successes           int64
	Failures            int64
	ConsecutiveFailures int64
	LastLatency         time.Duration
	LastSuccess         time.Time
	LastError           string
	CooldownUntil       time.Time
}

// HealthStore records provider success/failure health in a concurrency-safe way.
type HealthStore struct {
	mu                sync.RWMutex
	byProvider        map[string]HealthSnapshot
	cooldownThreshold int64
	cooldownDuration  time.Duration
	now               func() time.Time
}

// NewHealthStore constructs a provider health store.
func NewHealthStore(cooldownThreshold int64, cooldownDuration time.Duration) *HealthStore {
	if cooldownThreshold < 1 {
		cooldownThreshold = 1
	}
	if cooldownDuration < 0 {
		cooldownDuration = 0
	}
	return &HealthStore{
		byProvider:        make(map[string]HealthSnapshot),
		cooldownThreshold: cooldownThreshold,
		cooldownDuration:  cooldownDuration,
		now:               time.Now,
	}
}

// RecordSuccess records a successful provider attempt and clears consecutive
// failure/cooldown state for that provider.
func (s *HealthStore) RecordSuccess(provider string, latency time.Duration) {
	if s == nil || provider == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.byProvider[provider]
	snap.Provider = provider
	snap.Successes++
	snap.ConsecutiveFailures = 0
	snap.LastLatency = latency
	snap.LastSuccess = s.now()
	snap.LastError = ""
	snap.CooldownUntil = time.Time{}
	s.byProvider[provider] = snap
}

// RecordFailure records a failed provider attempt and enters cooldown once the
// configured consecutive-failure threshold is reached.
func (s *HealthStore) RecordFailure(provider string, latency time.Duration, err error) {
	if s == nil || provider == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.byProvider[provider]
	snap.Provider = provider
	snap.Failures++
	snap.ConsecutiveFailures++
	snap.LastLatency = latency
	if err != nil {
		snap.LastError = err.Error()
	} else {
		snap.LastError = ""
	}
	if snap.ConsecutiveFailures >= s.cooldownThreshold && s.cooldownDuration > 0 {
		snap.CooldownUntil = s.now().Add(s.cooldownDuration)
	}
	s.byProvider[provider] = snap
}

// Snapshot returns deterministic copies of all provider health records.
func (s *HealthStore) Snapshot() []HealthSnapshot {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]HealthSnapshot, 0, len(s.byProvider))
	for _, snap := range s.byProvider {
		out = append(out, snap)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out
}
