# gollmfree

> Status: M0–M6 complete / pre-v0.1.0. Core library, PollinationsAI provider, selector/health/fallback, three additional provider stubs (Chatai/Yqcloud/WeWordle), and CLI are implemented and tested. gormes-agent integration is live. M7 release hardening is next.

`gollmfree` is a pure Go library and CLI for routing chat-completion requests to anonymous/free LLM providers without API keys, sign-up, browser automation, Docker, a server process, or extra infrastructure.

This project is a partial Go port/reference implementation based on [`xtekky/gpt4free`](https://github.com/xtekky/gpt4free). Provider behavior, model aliases, request shaping, and fallback ideas are studied from upstream before each implementation slice, then adapted into tested Go code.

## Install

```bash
go get github.com/TrebuchetDynamics/gollmfree
```

```bash
go install github.com/TrebuchetDynamics/gollmfree/cmd/gollmfree@latest
```

## Go quick start

```go
import (
    "context"
    "fmt"

    "github.com/TrebuchetDynamics/gollmfree"
    "github.com/TrebuchetDynamics/gollmfree/providers"
)

func main() {
    poll := providers.NewPollinationsAI()
    registry, _ := gollmfree.NewRegistry(gollmfree.ProviderInfo{
        Name:            poll.Name(),
        Provider:        poll,
        SupportedModels: poll.SupportedModels(),
        DefaultPriority: 1,
    })
    client := gollmfree.NewClient(
        gollmfree.WithRegistry(registry),
        gollmfree.WithTimeout(30 * time.Second),
        gollmfree.WithMaxRetries(0),
    )
    resp, err := client.ChatCompletion(context.Background(), gollmfree.ChatRequest{
        Model:    "auto",
        Messages: []gollmfree.Message{{Role: "user", Content: "Hello"}},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Choices[0].Message.Content)
}
```

## CLI quick start

```bash
# Send a message
gollmfree chat "what is 2+2?"

# Stream output
gollmfree chat --stream "explain Go interfaces"

# List registered providers and health
gollmfree list

# Show model aliases and provider coverage
gollmfree models
```

## Provider status

| Provider | Status | Notes |
| --- | --- | --- |
| PollinationsAI | **active** | No-auth OpenAI-shaped endpoint. Rate limit: 1 concurrent anonymous request per IP queue. Default model `openai-fast`. Live smoke test: `GOLLMFREE_POLLINATIONS_LIVE=1 go test -tags=integration ./providers -run TestPollinationsAILiveSmoke`. |
| Chatai | **inactive** | Code implemented (`providers/chatai.go`), `httptest` suite passes. Live endpoint `chatai.ren` DNS SERVFAIL as of 2026-06-21. Re-activate when a working endpoint is confirmed. |
| Yqcloud | **inactive** | Code implemented (`providers/yqcloud.go`), `httptest` suite passes. `chat9.yqcloud.top` resolves but returns 405 on all POST paths as of 2026-06-21. |
| WeWordle | **inactive** | Code implemented (`providers/wewordle.go`), `httptest` suite passes. `wewordle.org/gptapi/v1/en/trial` returns 404 as of 2026-06-21. |
| DeepAI | postponed | Absent from upstream `gpt4free` at commit `798d8586`. |
| You.com | postponed | `needs_auth = True`, requires cookies/browser; out of no-auth scope. |
| LambdaChat | postponed | Upstream `working = False`, multi-step cookie/form flow. |

Upstream reference: <https://github.com/xtekky/gpt4free> at commit `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca`.

## Wiring multiple providers

```go
poll := providers.NewPollinationsAI()
chatai := providers.NewChatai()   // inactive live endpoint — will fail over
registry, _ := gollmfree.NewRegistry(
    gollmfree.ProviderInfo{Name: poll.Name(),   Provider: poll,   SupportedModels: poll.SupportedModels(),   DefaultPriority: 1},
    gollmfree.ProviderInfo{Name: chatai.Name(), Provider: chatai, SupportedModels: chatai.SupportedModels(), DefaultPriority: 2},
)
client := gollmfree.NewClient(
    gollmfree.WithRegistry(registry),
    gollmfree.WithMaxRetries(0), // fail fast; selector ranks healthy providers first
)
```

The selector ranks providers by consecutive failures and cooldown state. Providers that return errors are deprioritised automatically; after `maxConsecutiveFailures` they enter a 5-minute cooldown window.

## Testing policy

Normal tests use `httptest` and must not depend on live providers. Live smoke tests require an explicit opt-in:

```bash
# All tests (no network)
go test ./...

# PollinationsAI live smoke
GOLLMFREE_POLLINATIONS_LIVE=1 go test -tags=integration ./providers -run TestPollinationsAILiveSmoke
```

## Privacy and provider caveats

Prompts are sent to third-party anonymous providers that this project does not control. Do not send secrets, credentials, private data, or sensitive production information. Provider labels are treated as provider claims, not guarantees about the actual underlying model.

`gollmfree` does not collect credentials, require account creation, run a local daemon, or log prompt contents by default.

## Upstream reference

- Repository: <https://github.com/xtekky/gpt4free>
- Baseline commit: `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca`
- License: upstream is GNU GPL v3.0. Attribution and no-vendoring policy: [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

## Project plan

[`GOLLMFREE-PRD.md`](GOLLMFREE-PRD.md) is the master project file. It tracks architecture, roadmap, TDD evidence, blockers, decisions, and the v0.1.0 definition of done.
