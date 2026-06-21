//go:build integration

package providers

import (
	"context"
	"os"
	"testing"
	"time"

	gollmfree "github.com/TrebuchetDynamics/gollmfree"
)

func TestPollinationsAILiveSmoke(t *testing.T) {
	if os.Getenv("GOLLMFREE_POLLINATIONS_LIVE") != "1" {
		t.Skip("set GOLLMFREE_POLLINATIONS_LIVE=1 to run live PollinationsAI smoke test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := NewPollinationsAI().Complete(ctx, []gollmfree.Message{{Role: "user", Content: "Reply with one short sentence."}})
	if err != nil {
		t.Fatalf("live Complete returned error: %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		t.Fatalf("live Complete returned empty response: %#v", resp)
	}
}
