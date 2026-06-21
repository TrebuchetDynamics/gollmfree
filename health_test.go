package gollmfree

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestHealthStoreRecordsSuccessFailureAndCooldown(t *testing.T) {
	store := NewHealthStore(2, time.Minute)
	boom := errors.New("boom")

	store.RecordFailure("PollinationsAI", 120*time.Millisecond, boom)
	snap := store.Snapshot()[0]
	if snap.Provider != "PollinationsAI" || snap.Failures != 1 || snap.ConsecutiveFailures != 1 || snap.LastError != "boom" {
		t.Fatalf("failure snapshot = %#v", snap)
	}
	if !snap.CooldownUntil.IsZero() {
		t.Fatalf("first failure entered cooldown unexpectedly: %s", snap.CooldownUntil)
	}

	store.RecordFailure("PollinationsAI", 150*time.Millisecond, boom)
	snap = store.Snapshot()[0]
	if snap.ConsecutiveFailures != 2 || snap.CooldownUntil.IsZero() {
		t.Fatalf("second failure snapshot = %#v, want cooldown at threshold", snap)
	}

	store.RecordSuccess("PollinationsAI", 80*time.Millisecond)
	snap = store.Snapshot()[0]
	if snap.Successes != 1 || snap.ConsecutiveFailures != 0 || !snap.CooldownUntil.IsZero() || snap.LastLatency != 80*time.Millisecond {
		t.Fatalf("success snapshot = %#v, want reset consecutive failures/cooldown", snap)
	}
}

func TestHealthStoreConcurrencySafeSnapshots(t *testing.T) {
	store := NewHealthStore(3, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.RecordSuccess("pollinationsai", time.Millisecond)
			store.RecordFailure("pollinationsai", 2*time.Millisecond, errors.New("transient"))
			_ = store.Snapshot()
		}()
	}
	wg.Wait()
	snaps := store.Snapshot()
	if len(snaps) != 1 {
		t.Fatalf("snapshot length = %d, want 1", len(snaps))
	}
	if snaps[0].Successes != 50 || snaps[0].Failures != 50 {
		t.Fatalf("snapshot counts = successes %d failures %d, want 50/50", snaps[0].Successes, snaps[0].Failures)
	}
}
