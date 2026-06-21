# gollmfree

> Status: first-provider + selector/fallback scaffold / pre-v0.1.0. Core API/test harness exists, PollinationsAI has mocked request/response/error/stream coverage plus an opt-in live smoke test, and concurrency-safe health tracking, deterministic selector ranking, sequential fallback, per-attempt timeouts, retries, and race mode exist. Provider portfolio expansion and CLI behavior are still being built through strict TDD from [`GOLLMFREE-PRD.md`](GOLLMFREE-PRD.md).

`gollmfree` is a pure Go library and CLI for routing chat-completion requests to currently available anonymous/free LLM providers without API keys, sign-up, browser automation, Docker, a server process, or extra infrastructure.

This project is a partial Go port/reference implementation based on [`xtekky/gpt4free`](https://github.com/xtekky/gpt4free). Provider behavior, model aliases, request shaping, and fallback ideas should be studied from upstream before each implementation slice, then adapted into tested Go code rather than vendoring Python source.

## Install

The intended module path is:

```bash
go get github.com/TrebuchetDynamics/gollmfree
```

The CLI target will be installable as:

```bash
go install github.com/TrebuchetDynamics/gollmfree/cmd/gollmfree@latest
```

> These commands may not work until the implementation scaffold and CLI are complete.

## Go quick start

Implemented public data types: `Message`, `ChatRequest`, `CompletionResponse`, `Choice`, and `StreamChunk`. The provider contract is also available as `Provider`, with immutable registry support through `NewRegistry`, `Registry`, `ProviderInfo`, and `ModelInfo`. `NewClient`, option helpers (`WithTimeout`, `WithMaxRetries`, `WithRaceMode`, `WithRaceWidth`, `WithProviderPriority`), `HealthStore` snapshots, selector ranking, sequential `ChatCompletion` fallback, per-attempt timeouts, retries, and race orchestration are implemented.

Planned chat API shape, pending provider and selector work:

```go
client := gollmfree.NewClient()
resp, err := client.ChatCompletion(ctx, gollmfree.ChatRequest{
    Model: "auto",
    Messages: []gollmfree.Message{{Role: "user", Content: "Hello"}},
})
if err != nil {
    return err
}
fmt.Println(resp.Choices[0].Message.Content)
```

## CLI quick start

Planned commands:

```bash
gollmfree chat "hello"
gollmfree chat --stream "hello"
gollmfree list
gollmfree models
```

## Upstream reference

Canonical upstream reference:

- Repository: <https://github.com/xtekky/gpt4free>
- Baseline inspected commit: `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca`
- Role: source reference for provider behavior, provider metadata, request formats, parsing behavior, streaming behavior, and fallback strategy.
- Porting policy: inspect current upstream source and commit SHA for every provider slice; document what was ported, changed, omitted, or postponed.
- License note: upstream `LICENSE` is GNU GPL v3.0 and `LEGAL_NOTICE.md` disclaims provider affiliation/warranty. Attribution and no-vendoring policy are tracked in [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md); legal compatibility remains a release review item.

## Provider status

| Provider | Status | Upstream reference | Notes |
| --- | --- | --- | --- |
| PollinationsAI | implemented-untested-live | `g4f/Provider/PollinationsAI.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Current upstream, no-auth text endpoint `https://text.pollinations.ai/openai`, default model `openai-fast`; non-streaming request/response, error handling, and stream emulation have mocked coverage; live smoke is opt-in with `GOLLMFREE_POLLINATIONS_LIVE=1 go test -tags=integration ./providers -run TestPollinationsAILiveSmoke`. |
| Chatai | selected next | `g4f/Provider/Chatai.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Current upstream replacement for legacy ChatgptAi; `working = True`, `needs_auth = False`, SSE endpoint `https://chatai.aritek.app/stream`; mocked provider next. |
| Yqcloud | viable candidate | `g4f/Provider/Yqcloud.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Current upstream, no-auth stream endpoint `https://api.binjie.fun/api/generateStream`; implement after Chatai if still suitable. |
| WeWordle | viable candidate | `g4f/Provider/WeWordle.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Current upstream, no-auth endpoint `https://wewordle.org/gptapi/v1/web/turbo`; retry/raw-or-JSON stream behavior needs mocked port if chosen. |
| DeepAI | postponed/replaced | upstream commit `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` has no DeepAI provider/model metadata | Revisit only if endpoint is independently revalidated. |
| ChatgptAi | absent upstream/postponed | no current upstream provider file at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Replaced by Chatai for v0.1.0 planning. |
| ChatgptLogin | absent upstream/postponed | no current upstream provider file at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Postponed. |
| Ails | absent upstream/postponed | no current upstream provider file at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Postponed. |
| You.com | postponed | `g4f/Provider/needs_auth/You.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | `needs_auth = True` and may require cookies/browser automation; out of v0.1.0 no-auth scope. |
| LambdaChat | postponed | `g4f/Provider/LambdaChat.py` at `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca` | Upstream marks `working = False` and flow is multi-step cookie/conversation/form handling. |

## Testing policy

Development follows strict TDD:

1. write one failing behavior test;
2. implement the smallest code to pass;
3. refactor while tests stay green;
4. update `GOLLMFREE-PRD.md` and this README when user-facing behavior or status changes.

Normal tests must not depend on live anonymous providers. Provider implementations should use `httptest` for deterministic coverage. Live provider smoke tests must be opt-in with an integration build tag.

## Privacy and provider caveats

Prompts are sent to third-party anonymous providers that this project does not control. Do not send secrets, credentials, private data, or sensitive production information. Provider labels are treated as provider claims, not guarantees about the actual underlying model.

`gollmfree` must not collect credentials, require OAuth/account creation, run a local daemon, or log prompt contents by default.

## Project plan

`GOLLMFREE-PRD.md` is the master project file. It tracks architecture, roadmap, TDD evidence, blockers, decisions, and the v0.1.0 definition of done.
