# Gollmfree Product Requirements Document

> **Project:** `gollmfree`  
> **Type:** Pure Go library + CLI  
> **Proposed module path:** `github.com/TrebuchetDynamics/gollmfree`  
> **Inspiration:** Python [`xtekky/gpt4free`](https://github.com/xtekky/gpt4free) provider strategy  
> **Primary consumer:** `gormes-agent`  
> **Version target:** v0.1.0
> **Planning status:** Master project file; update this document before and after every implementation slice.
> **Development mode:** Strict TDD. No production implementation without a failing test or documented spike exception.

## 0. Master Project Control

This file is the source of truth for creating the whole `gollmfree` project and tracking progress. If implementation reality changes, update this PRD in the same change as the code.

### 0.1 How to Use This File

Before starting any work slice:

1. Read sections 0, 8, 9, 11, 15, 16, and 17.
2. Pick exactly one unchecked task from the Master Progress Tracker or Detailed Backlog.
3. Confirm its prerequisite tasks are complete or explicitly blocked.
4. Write or update the test named in the task before production code.
5. Run the smallest relevant failing test and record the red result in this file if useful.
6. Implement the smallest code change to pass.
7. Run focused tests, then the milestone gate command.
8. Update status, evidence, and next task in this file.

### 0.2 Progress Status Legend

| Status | Meaning |
| --- | --- |
| `todo` | Not started. |
| `red` | Failing test exists and describes desired behavior. |
| `green` | Focused tests pass for the slice. |
| `refactored` | Code was cleaned after green while tests stayed passing. |
| `blocked` | Cannot continue without a decision or external evidence. |
| `done` | Done signal is met and evidence is recorded. |

### 0.3 Master Progress Tracker

Keep this table current. Do not mark a milestone `done` unless every task in that milestone has evidence.

| Milestone | Status | Current task | Required evidence before `done` |
| --- | --- | --- | --- |
| M0 Decisions and Scaffold | done | T1.1 Public types | Final module path chosen; `go.mod` created; living README exists; CI skeleton exists; `go test ./...` runs. |
| M1 Core API and Test Harness | todo | T1.1 Public types | Public API types, provider interface, registry, options, fake-provider harness, tests passing. |
| M2 First Provider Vertical Slice | todo | T2.1 DeepAI endpoint research | DeepAI mocked tests pass; integration smoke test is opt-in and skipped by default. |
| M3 Selector, Health, and Fallback | todo | T3.1 Health store | Deterministic tests for ranking, fallback, cooldown, retries, race mode, concurrency. |
| M4 Provider Portfolio | todo | T4.1 Provider viability pass | At least three providers implemented or explicitly status-documented with tests/stubs. |
| M5 CLI | todo | T5.1 CLI command router | `go install ./cmd/gollmfree`; `list`/`models` no-network tests; `chat`/stream use client path. |
| M6 `gormes-agent` Integration | todo | T6.1 Inspect `gormes-agent` | Actual interface inspected; compiling example/adapter or documented blocker. |
| M7 Release Hardening | todo | T7.1 README quick start | README, examples, GoDoc, CI, `go test ./...`, `go vet ./...`, DoD checklist complete. |

### 0.4 Current Next Action

- **Next task:** T1.1 Public types.
- **Reason:** M0 scaffold is complete; the first implementation slice is defining the public request/response types through a failing compile/example test.
- **First test/evidence:** A focused test/example should fail until `types.go` defines `Message`, `ChatRequest`, `CompletionResponse`, `Choice`, and `StreamChunk`.

### 0.5 Decision Log

| Date | Decision | Reason | Revisit trigger |
| --- | --- | --- | --- |
| 2026-06-09 | Use `github.com/TrebuchetDynamics/gollmfree` as the module path. | Public GitHub repository `TrebuchetDynamics/gollmfree` was created. | If ownership/repository path changes before release. |
| TBD | Use `providers/` subpackage for provider implementations. | Keeps provider fragility local and matches maintainability goal. | If import cycles or registration ergonomics become poor. |
| TBD | Race mode disabled by default. | Avoids extra traffic to anonymous providers. | If sequential mode success latency is unacceptable. |
| TBD | CLI uses standard `flag` for v0.1.0. | Keeps dependencies minimal. | If command UX becomes complex enough to justify Cobra. |

### 0.6 Blocker Log

| Blocker | Status | Owner action needed | Unblocks |
| --- | --- | --- | --- |
| Final module owner path unknown | resolved | Module path chosen: `github.com/TrebuchetDynamics/gollmfree`. | T0.1, install docs, examples. |
| `gormes-agent` LLM interface unknown | open | Inspect repository/interface before M6. | T6.1, T6.2. |
| Live provider viability unknown | open | Re-check endpoints during provider viability pass. | T2.1, T4.1-T4.4. |

### 0.7 Evidence Log

| Date | Task | Red | Green/Gate | Files | Next |
| --- | --- | --- | --- | --- | --- |
| 2026-06-09 | T0.1 Choose module path | Documentation/scaffold task; no production behavior test required. | `go list -m` -> `github.com/TrebuchetDynamics/gollmfree` | `GOLLMFREE-PRD.md`, `go.mod` | T0.2/T0.4 |
| 2026-06-09 | T0.4 Create living README skeleton | Documentation task; no production behavior test required. | README created with status, install path, planned API/CLI, provider table, testing policy, privacy caveats. | `README.md`, `GOLLMFREE-PRD.md` | T0.2/T0.3 |
| 2026-06-09 | T0.2 Scaffold module | Scaffold task; no production behavior test required. | `go test ./...`; `git diff --check` | `go.mod`, `doc.go`, `GOLLMFREE-PRD.md` | T0.3 |
| 2026-06-09 | T0.3 Add CI skeleton | CI configuration task; no production behavior test required. | `.github/workflows/test.yml` runs `go test ./...` and `go vet ./...`; local `go test ./...`; local `go vet ./...`; `git diff --check` | `.github/workflows/test.yml`, `GOLLMFREE-PRD.md` | T1.1 |

### 0.8 Update Discipline

Every implementation PR or agent session must update at least one of:

- Master Progress Tracker status/current task.
- Detailed Backlog row done signal/evidence.
- Decision Log.
- Blocker Log.
- Known Gaps.
- Definition of Done checklist.

Every milestone completion must also update `README.md` if user-facing behavior, install steps, provider status, caveats, examples, or project status changed. The README is a living companion to this PRD, not a release-only artifact.

If no PRD or README update is needed, the session must explicitly say why.

## 1. Vision

Gollmfree is a drop-in Go package that lets Go programs talk to the best currently available free LLM provider without API keys, user sign-up, browser automation, a server process, Docker, or extra infrastructure.

The library hides provider fragility behind a ranked selector that tracks health, latency, failures, cooldowns, retries, and fallback. Developers should be able to import one package, create one client, and receive an OpenAI-shaped response from whichever anonymous/free provider is currently working.

The CLI exposes the same capability from the terminal for immediate manual use and smoke testing.

## 2. Goals and Non-Goals

### 2.1 Goals

- Provide a pure Go client library with a zero-configuration default.
- Support synchronous and streaming chat-completion APIs.
- Port a small set of simple, reliable `g4f` providers first.
- Rank providers automatically by explicit priority and dynamic health.
- Fall back across providers when one fails.
- Offer a CLI binary named `gollmfree` using the same library path.
- Make integration into `gormes-agent` require minimal code changes.
- Keep dependencies minimal and avoid CGO, browser automation, servers, GUI, Docker, and hosted infrastructure.

### 2.2 Non-Goals

- No hosted API gateway.
- No OpenAI-compatible HTTP server in v0.1.0.
- No account creation, user credential management, paid-key management, or OAuth.
- No browser automation or headless browser dependencies for initial providers.
- No guarantee that upstream anonymous providers are permanently available.
- No promise that provider-advertised model labels are authentic; labels are treated as provider claims.

## 3. Target Users

1. **Developer integrating with `gormes-agent`**  
   Wants a zero-cost LLM backend that can be swapped into the agent framework quickly.

2. **Go developers needing a free LLM client**  
   Want an API that feels close enough to OpenAI chat completions but does not require keys or billing.

3. **CLI users and maintainers**  
   Want to test provider availability and get quick completions from a terminal.

## 4. Key User Stories

- As a Go developer, I can call `gollmfree.NewClient()` with no configuration and get a response.
- As a Go developer, I can ask for `gpt-3.5-turbo`, `deepai`, `auto`, or `best` and let the library choose a working provider.
- As a Go developer, I can stream chunks when a provider supports streaming or receive an emulated one-shot stream otherwise.
- As a maintainer, I can add a new provider in one file and register supported models and priority.
- As a CLI user, I can run `gollmfree chat "hello"` and receive an answer.
- As a CLI user, I can run `gollmfree list` and see provider status.
- As a `gormes-agent` integrator, I can replace the existing LLM client with `gollmfree.NewClient()` or a small adapter.

## 5. Functional Requirements

### FR-1: Provider Interface

Every provider MUST implement one common interface.

```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, messages []Message) (CompletionResponse, error)
    Stream(ctx context.Context, messages []Message) (<-chan string, error)
    SupportedModels() []string
}
```

Requirements:

- `Complete` MUST respect `context.Context` cancellation and deadlines.
- `Stream` MUST return a channel of text chunks and close it when complete.
- Providers that do not support native streaming MAY emit the full response as a single chunk.
- Provider implementations MUST NOT require user-supplied auth parameters.
- Anonymous tokens, public headers, scraping headers, and provider-specific request details are encapsulated inside each provider.
- Provider errors MUST include the provider name and enough context for combined fallback errors.

### FR-2: Provider Implementations and Registry

Initial providers SHOULD be ported from the simpler and more reliable `g4f` providers:

| Priority | Provider | Go file | Notes |
| --- | --- | --- | --- |
| 1 | DeepAI | `providers/deepai.go` or `deepai.go` | First implementation target; non-streaming may emulate stream. |
| 2 | ChatgptAi | `providers/chatgptai.go` | Scrapes `chatgpt.ai`; validate current endpoint before implementation. |
| 3 | Yqcloud | `providers/yqcloud.go` | Uses `chat9.yqcloud.top` if still active. |
| 4 | ChatgptLogin | `providers/chatgptlogin.go` | Potentially slow; lower default priority. |
| 5 | Ails | `providers/ails.go` | Include only if still active and not too complex. |
| 6 | You.com | `providers/you.go` | Postpone if headers/session flow is complex. |

Registry requirements:

- The library MUST expose a registry of available providers.
- Each registry entry MUST include provider name, supported model aliases, and default priority.
- Registry operations MUST be safe for concurrent reads.
- Adding a provider SHOULD require only implementing the interface and registering metadata.

Example registry shape:

```go
type ProviderInfo struct {
    Name            string
    Provider        Provider
    SupportedModels []string
    DefaultPriority int
}
```

### FR-3: Ranked Automated Selector

The selector is the core differentiator. Users request a model; the selector chooses the best provider.

Requirements:

- Accepted model strings MUST include at least:
  - `auto`
  - `best`
  - provider-specific names such as `deepai`
  - common aliases such as `gpt-3.5-turbo` when providers claim support
- The selector MUST maintain ordered candidate lists per model.
- Ordering MUST combine:
  - explicit default priorities,
  - user priority overrides,
  - dynamic health score,
  - recent latency,
  - cooldown state after repeated failures.
- On a normal completion, the selector MUST attempt providers in ranked order until one succeeds.
- Each provider attempt MUST use a configurable per-attempt timeout.
- On success, the selector MUST record success and latency.
- On failure, the selector MUST record failure and try the next candidate.
- If all candidates fail, the selector MUST return a combined error containing each provider failure.
- The selector SHOULD support a race mode that calls multiple providers concurrently and returns the first successful response.
- Ranking logic MUST remain internal to the library; callers use `Client.ChatCompletion` or `Client.ChatCompletionStream`.

### FR-4: Client API

The package MUST expose a zero-configuration client.

```go
client := gollmfree.NewClient()
resp, err := client.ChatCompletion(ctx, gollmfree.ChatRequest{
    Model: "gpt-3.5-turbo",
    Messages: []gollmfree.Message{
        {Role: "user", Content: "Hello"},
    },
})
if err != nil {
    return err
}
fmt.Println(resp.Choices[0].Message.Content)
```

Streaming variant:

```go
stream, err := client.ChatCompletionStream(ctx, req)
if err != nil {
    return err
}
for chunk := range stream {
    fmt.Print(chunk.Content)
}
```

Minimum public types:

```go
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Stream      bool      `json:"stream,omitempty"`
    Temperature *float64  `json:"temperature,omitempty"`
    MaxTokens   *int      `json:"max_tokens,omitempty"`
}

type CompletionResponse struct {
    ID       string   `json:"id,omitempty"`
    Object   string   `json:"object,omitempty"`
    Created  int64    `json:"created,omitempty"`
    Model    string   `json:"model,omitempty"`
    Provider string   `json:"provider,omitempty"`
    Choices  []Choice `json:"choices"`
}

type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason,omitempty"`
}

type StreamChunk struct {
    Content  string `json:"content"`
    Provider string `json:"provider,omitempty"`
    Model    string `json:"model,omitempty"`
}
```

Configuration options:

```go
client := gollmfree.NewClient(
    gollmfree.WithTimeout(15*time.Second),
    gollmfree.WithMaxRetries(2),
    gollmfree.WithProviderPriority("gpt-3.5-turbo", []string{"deepai", "yqcloud"}),
    gollmfree.WithRaceMode(false),
)
```

### FR-5: CLI Tool

A single binary named `gollmfree` MUST be provided.

Required commands:

```bash
gollmfree chat "your prompt"
gollmfree chat --stream "your prompt"
gollmfree list
gollmfree models
```

CLI requirements:

- `chat` MUST print the completion to stdout.
- `chat --stream` MUST print chunks as they arrive.
- `list` MUST show providers and health/status.
- `models` MUST show known model aliases and provider support.
- The CLI MUST use the same library API as Go consumers.
- The CLI SHOULD use only the standard `flag` package for v0.1 unless a dependency is strongly justified.
- Users SHOULD be able to install with:

```bash
go install github.com/TrebuchetDynamics/gollmfree/cmd/gollmfree@latest
```

### FR-6: `gormes-agent` Integration

Integration requirements:

- The module MUST be importable as `github.com/TrebuchetDynamics/gollmfree` once the final owner path is chosen.
- No CGO, browser automation, Docker, server process, or local daemon may be required.
- Provider selection and failover MUST be hidden from `gormes-agent`.
- The library MUST expose enough OpenAI-shaped response data for a small adapter if `gormes-agent` has its own LLM interface.
- Documentation MUST include a `gormes-agent` integration example.

Example target shape:

```go
agent := gormes.NewAgent(
    gormes.WithLLM(gollmfree.NewClient()),
)
```

If `gormes-agent` requires a custom interface, provide an adapter example in `examples/gormes-agent/`.

### FR-7: No Server, GUI, or Docker

- Gollmfree MUST run in-process.
- The CLI MUST be a thin wrapper around the library.
- No HTTP server, GUI, Docker image, or daemon is required for v0.1.0.
- Documentation MUST focus on Go module consumption and CLI usage.

## 6. Non-Functional Requirements

### Performance

- Selector overhead SHOULD remain under 200ms beyond upstream provider latency for sequential mode.
- Race mode MAY use more network calls to reduce wall-clock latency.
- Provider ranking and health lookup SHOULD be O(n) over candidate providers for the requested model.

### Concurrency

- `Client` MUST be safe for concurrent use by multiple goroutines.
- Health data MUST be protected by mutexes or other safe synchronization.
- Registry reads MUST be safe during normal use.

### Resilience

- Provider attempts MUST use context deadlines.
- Retries SHOULD use bounded backoff.
- Repeated failures SHOULD place a provider in cooldown.
- Cooldown SHOULD expire automatically so providers can recover.
- Combined errors MUST expose all attempted provider failures.

### Maintainability

- Each provider SHOULD live in a separate file.
- Provider request/response parsing SHOULD be covered by mocked unit tests where possible.
- Integration tests against real providers MUST be skippable in short/CI mode.
- New providers SHOULD not require selector changes.

### Go Idioms and Dependencies

- Use `context.Context` everywhere network calls are made.
- Use `net/http`, `encoding/json`, `errors`, `sync`, and standard library primitives first.
- Avoid CGO.
- Avoid large dependency trees in v0.1.0.
- Optional dependencies must be justified in the README.

## 7. Proposed Package Layout

```text
.
├── go.mod
├── client.go
├── types.go
├── provider.go
├── registry.go
├── selector.go
├── health.go
├── errors.go
├── providers/
│   ├── deepai.go
│   ├── chatgptai.go
│   ├── yqcloud.go
│   ├── chatgptlogin.go
│   └── ails.go
├── cmd/
│   └── gollmfree/
│       └── main.go
├── examples/
│   ├── basic/
│   ├── streaming/
│   └── gormes-agent/
├── README.md
└── .github/
    └── workflows/
        └── test.yml
```

## 8. Architecture Plan

### 8.1 Architectural Principles

- Keep the public API small: callers interact with `Client`, `ChatRequest`, response types, and functional options.
- Keep provider fragility local: endpoint URLs, headers, request payloads, response parsing, and quirks live inside provider implementations.
- Make provider choice an internal concern: ranking, cooldowns, retries, fallbacks, and race mode stay behind the client API.
- Prefer standard-library dependencies for v0.1.0.
- Treat tests as the main stability mechanism because external anonymous providers are unreliable.

### 8.2 Module Responsibilities

| Module | Files | Responsibility | Public Surface |
| --- | --- | --- | --- |
| Public client | `client.go`, `options.go` | Owns caller-facing chat and stream methods, validates requests, delegates selection, maps results to OpenAI-shaped responses. | `NewClient`, `Client.ChatCompletion`, `Client.ChatCompletionStream`, `Option` functions. |
| Types | `types.go` | Defines stable request/response data structures. | `Message`, `ChatRequest`, `CompletionResponse`, `Choice`, `StreamChunk`. |
| Provider contract | `provider.go` | Defines the narrow interface all providers satisfy. | `Provider`, optional capability/status helpers if needed. |
| Registry | `registry.go` | Stores provider metadata, aliases, priorities, and available providers. | Read-only listing functions and registration helpers. |
| Selector | `selector.go` | Builds candidate lists by model and chooses attempt order from priority plus health. | Internal to library. |
| Health tracking | `health.go` | Records successes, failures, latency, last error, and cooldown state in a concurrency-safe store. | Internal snapshots for CLI status. |
| Attempts/errors | `attempt.go`, `errors.go` | Applies per-attempt timeout/retry policy and returns combined fallback errors. | Error types safe for callers to inspect/log. |
| Providers | `providers/*.go` | Encapsulates endpoint-specific requests, parsing, streaming emulation, and provider errors. | Provider constructors or registered defaults. |
| CLI | `cmd/gollmfree/main.go` | Thin terminal wrapper around the same client and registry APIs. | Commands only; no separate business logic. |
| Examples/docs | `examples/*`, `README.md` | Demonstrate Go API, streaming, CLI, and `gormes-agent` integration. | N/A. |

### 8.3 Request Flow

```text
Caller / CLI
    |
    v
Client.ChatCompletion / Client.ChatCompletionStream
    |
    | validate request, apply defaults, derive model alias
    v
Selector candidates(model)
    |
    | registry aliases + user priority overrides + health snapshot
    v
Attempt loop or race mode
    |
    | per-provider timeout, bounded retry, context cancellation
    v
Provider.Complete / Provider.Stream
    |
    | provider-specific HTTP and parsing
    v
Health recorder + response mapper
    |
    v
CompletionResponse / StreamChunk channel / combined error
```

### 8.4 Package and Dependency Direction

```text
cmd/gollmfree  --->  gollmfree public client
examples/*     --->  gollmfree public client
client.go      --->  selector, registry, health, providers through Provider interface
selector.go    --->  registry metadata + health snapshots
providers/*    --->  provider contract + stdlib net/http/json
registry.go    --->  provider contract only
health.go      --->  stdlib sync/time only
```

Rules:

- Providers MUST NOT import the selector or client.
- The selector MUST use the `Provider` interface and metadata, not concrete provider types.
- The CLI MUST NOT call providers directly.
- Health state MUST be updated only by the client/attempt path.
- Registry default construction SHOULD happen in one place so tests can inject fake registries.

### 8.5 Core Data Structures

Recommended internal shapes:

```go
type Client struct {
    registry *Registry
    selector *Selector
    health   *HealthStore
    http     *http.Client
    options  clientOptions
}

type Registry struct {
    providers []ProviderInfo
    byName    map[string]ProviderInfo
    byModel   map[string][]ProviderInfo
}

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
```

### 8.6 Ranking Strategy

Sequential ranking SHOULD sort candidates by:

1. Explicit user priority for the requested model.
2. Provider cooldown state; providers in active cooldown move to the end or are skipped until all healthy providers fail.
3. Default priority from `ProviderInfo`.
4. Consecutive failures and recent failure count.
5. Recent latency, preferring lower latency when health is otherwise similar.
6. Stable provider name as a final deterministic tie-breaker.

Race mode SHOULD reuse the same ranked list, start only the top N configurable candidates, cancel remaining attempts after the first success, and record health for every completed attempt.

### 8.7 Error Model

- Provider-specific errors should wrap the root cause and include provider name.
- Selector failure should return a combined error with attempted providers in order.
- Context cancellation and deadline errors should remain detectable with `errors.Is`.
- CLI output should print concise user-facing errors while verbose provider detail remains available in debug mode later.

### 8.8 Streaming Strategy

- Native provider streams should be adapted to `StreamChunk` values at the client layer.
- Non-streaming providers may emulate streaming by sending one chunk with the complete response.
- Stream channels must close exactly once.
- If an error occurs before streaming starts, return it from `ChatCompletionStream`.
- If an error occurs after streaming starts, v0.1.0 may close the channel and record health; a future version can add an error-bearing stream event type if needed.

### 8.9 Testing Architecture

- Use fake `Provider` implementations for selector, health, fallback, race mode, timeout, and concurrency tests.
- Use `httptest.Server` for provider request/response parsing tests.
- Keep real-provider integration tests opt-in with tags and/or environment variables.
- CLI tests should exercise argument parsing and command routing without depending on live providers.
- Race mode tests must use deterministic fake providers, not sleeps against real endpoints.

### 8.10 Documentation Architecture

The repository documentation should be split by consumer need:

- `README.md`: install, quick start, CLI use, privacy/provider caveats, current provider status, project maturity, and links to examples.
- `examples/basic`: minimal sync chat.
- `examples/streaming`: stream consumption.
- `examples/gormes-agent`: adapter or integration sketch after inspecting `gormes-agent`.
- GoDoc comments: public API contracts, concurrency guarantees, and error behavior.

`README.md` must be created during scaffolding and maintained periodically throughout development. It should never wait until release hardening. At minimum, update it when any of these change:

- module path or install command;
- public API shape or examples;
- CLI commands, flags, or exit behavior;
- provider implementation/status table;
- privacy, security, legal, or provider-instability caveats;
- `gormes-agent` integration status;
- validation commands or release maturity.

Recommended README sections from the start:

1. Project status and warning about anonymous providers.
2. Install / module path.
3. Go quick start.
4. CLI quick start.
5. Provider status table.
6. Streaming notes.
7. `gormes-agent` integration notes.
8. Testing and integration-test policy.
9. Privacy and acceptable-use caveats.

### 8.11 Public API Contract Details

`NewClient(opts ...Option)` should construct a concurrency-safe client with default registry, default selector, default health store, and default HTTP client. It must not perform network calls during construction.

Default option values for v0.1.0:

| Setting | Default | Reason |
| --- | ---: | --- |
| Model when empty | `auto` | Zero-config call path. |
| Per-attempt timeout | `15s` | Bounds broken providers without making normal free endpoints impossible. |
| Max retries | `0` or `1` | Prefer fallback over hammering one fragile endpoint. |
| Race mode | `false` | Avoid unnecessary upstream traffic by default. |
| Race width | `2` when race mode is enabled | Useful latency hedge without calling every provider. |
| Cooldown threshold | `3` consecutive failures | Avoid demoting a provider on one transient failure. |
| Cooldown duration | `5m` initial, capped at `30m` | Lets providers recover automatically. |

Client methods:

```go
func NewClient(opts ...Option) *Client
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (CompletionResponse, error)
func (c *Client) ChatCompletionStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
func (c *Client) Providers() []ProviderInfo
func (c *Client) Models() []ModelInfo
func (c *Client) Health() []HealthSnapshot
```

`Providers`, `Models`, and `Health` support the CLI without exposing selector internals. Returned slices must be copies so callers cannot mutate shared state.

### 8.12 Internal File Contracts

| File | Must contain | Must not contain |
| --- | --- | --- |
| `types.go` | JSON-shaped public request/response types and GoDoc comments. | Provider-specific fields or ranking state. |
| `provider.go` | `Provider` interface, `ProviderInfo`, optional `ModelInfo`. | Concrete provider registration side effects that make tests hard. |
| `registry.go` | Immutable registry construction, alias normalization, safe list/lookups. | Health scoring or network calls. |
| `selector.go` | Candidate filtering, deterministic ranking, race/sequential selection helpers. | HTTP request construction or provider-specific parsing. |
| `health.go` | Mutex-protected counters, cooldown calculations, snapshot copy methods. | Provider metadata or request validation. |
| `client.go` | Public methods, request defaults/validation, attempt orchestration. | Provider endpoint constants. |
| `options.go` | Functional option validation and internal config defaults. | Runtime mutation of registry or health. |
| `errors.go` | `ProviderError`, `AttemptError`, `CombinedError` and unwrap helpers. | CLI formatting policy. |
| `providers/*.go` | One provider per file with endpoint, request, parse, and tests. | Selector imports or global mutable health. |
| `cmd/gollmfree/main.go` | Command parsing, stdout/stderr behavior, exit codes. | Duplicate provider selection logic. |

### 8.13 Selector Algorithm Blueprint

Sequential mode should be implementable as this deterministic loop:

```text
normalize requested model; if empty use auto
candidates = registry.Candidates(model)
if candidates empty: return UnknownModelError
ranked = selector.Rank(candidates, health.Snapshot(), userOverrides)
errors = []AttemptError{}
for provider in ranked:
    if provider in cooldown and at least one non-cooldown candidate remains:
        continue until healthy candidates are tried
    for attempt = 0; attempt <= maxRetries; attempt++:
        attemptCtx = context.WithTimeout(parentCtx, perAttemptTimeout)
        start = now
        response, err = provider.Complete(attemptCtx, messages)
        latency = since(start)
        if err == nil:
            health.RecordSuccess(provider, latency)
            annotate response with provider and requested model
            return response
        health.RecordFailure(provider, latency, err)
        append AttemptError(provider, attempt, err)
        if parent context is canceled: return combined error preserving context error
return CombinedError(errors)
```

Race mode differences:

- Select the top `raceWidth` ranked candidates after cooldown filtering.
- Derive a child context; cancel it after first successful response.
- Record failures for completed failed attempts.
- Do not record a canceled loser as provider failure when cancellation was caused by another provider succeeding.
- If every raced provider fails, continue sequentially with remaining candidates or return the combined race errors; v0.1.0 should prefer continuing sequentially for higher success rate.

### 8.14 Model and Alias Plan

Model aliases are treated as routing hints, not authenticity guarantees.

| Alias | Meaning | Candidate behavior |
| --- | --- | --- |
| `auto` | Best available default. | All registered providers ordered by rank. |
| `best` | Same as `auto` for v0.1.0. | Reserved for future quality weighting. |
| Provider name, e.g. `deepai` | Force/prefer that provider. | That provider first; optionally fall back to others unless a future strict mode is added. |
| Claimed OpenAI label, e.g. `gpt-3.5-turbo` | Provider claims support or compatible response. | Providers whose metadata includes the alias. |

Alias normalization rules:

- Trim whitespace and lowercase model/provider keys.
- Preserve the caller-provided `ChatRequest.Model` in response metadata when useful, but expose actual provider name separately.
- Registry should reject duplicate provider names at construction time.
- Registry should allow multiple providers per alias.

### 8.15 Provider Implementation Playbook

Each provider should be added as a vertical slice:

1. Create `providers/<name>.go` with a small struct containing `http.Client`, endpoint URL, and optional test override URL.
2. Implement `Name`, `SupportedModels`, `Complete`, and `Stream`.
3. Build requests from `[]Message` using a helper that preserves role/content order.
4. Set only required public headers and document any copied browser-like headers in comments.
5. Parse successful responses into `CompletionResponse` with provider name set.
6. Convert non-2xx responses and malformed payloads into `ProviderError` with status/body snippet limits.
7. Respect `ctx` on every `http.NewRequestWithContext` call.
8. Emulate streaming by calling `Complete` and sending one chunk when native streaming is unavailable.
9. Add `httptest` unit tests for happy path, provider error, malformed response, timeout/cancel, and stream emulation.
10. Add integration test guarded by `//go:build integration` and environment variables if live smoke testing is useful.

Provider status should be documented in README as one of:

- `active`: implemented and recently smoke-tested.
- `implemented-untested-live`: mocked tests pass, live endpoint not verified recently.
- `inactive`: endpoint known broken but code kept for reference or disabled by default.
- `postponed`: not implemented because auth/session/browser flow is too complex for v0.1.0.

### 8.16 CLI Specification

All CLI commands must be thin wrappers around library APIs.

| Command | Required behavior | Network required? | Exit codes |
| --- | --- | --- | --- |
| `gollmfree chat "prompt"` | Send one user message and print final text plus trailing newline. | Yes | `0` success, `1` provider/client error, `2` usage error. |
| `gollmfree chat --stream "prompt"` | Print chunks as they arrive; add trailing newline if output did not end with one. | Yes | Same as `chat`. |
| `gollmfree list` | Print provider name, default priority, supported aliases, cooldown/health summary. | No | `0` success. |
| `gollmfree models` | Print aliases and provider coverage. | No | `0` success. |

Initial flags:

- `--model` default `auto`.
- `--timeout` default client timeout.
- `--race` default false.
- `--json` for `list` and `models` only if easy; otherwise postpone.

### 8.17 Observability and Privacy

- Do not log prompt contents by default.
- Error messages may include provider name, HTTP status, and a short sanitized body snippet, but not full prompts.
- Health snapshots should expose last error text for debugging; README must warn that provider errors can include upstream text.
- No telemetry, analytics, credential storage, or local server should be introduced in v0.1.0.

### 8.18 Implementation Risk Gates

Before moving from one milestone to the next:

- Foundation gate: public API compiles and fake-provider tests pass.
- Provider gate: at least one provider is tested through `httptest` without network.
- Selector gate: fallback/cooldown/race/concurrency behavior proven with deterministic fakes.
- CLI gate: `list` and `models` require no network and `chat` uses the client path.
- Integration gate: `gormes-agent` interface is inspected or its unavailability is documented.
- Release gate: `go test ./...`, `go vet ./...`, README, examples, and CI are complete.

## 9. Task Roadmap

This roadmap is ordered to reduce unknowns early: first lock the public seam, then prove one provider, then make fallback reliable, then expand the provider portfolio.

### 9.1 Roadmap Dependency Graph

```text
M0 decisions/scaffold
  -> M1 core API + fake-provider tests
      -> M2 first provider vertical slice
      -> M3 selector + health + fallback
          -> M4 provider portfolio
          -> M5 CLI
              -> M6 gormes-agent integration
                  -> M7 release hardening
```

Parallelization notes:

- M2 provider research can start while M1 tests are being finished, but provider code should wait for the `Provider` interface to stabilize.
- M5 CLI `list`/`models` can start after registry snapshots exist; `chat` should wait for the client path.
- M6 integration should not start until the client API is stable enough to avoid rewriting the example.

### 9.2 Work Item Format

Each implementation task should be trackable as a small issue with:

- Scope: files expected to change.
- Inputs: PRD requirement, provider endpoint notes, or external interface to inspect.
- Done signal: command/test/doc evidence that proves completion.
- Risk: live-provider instability, API uncertainty, concurrency risk, or docs-only risk.

### Milestone 0: Decisions and Scaffold

**Goal:** Remove release-blocking ambiguity before coding.

Tasks:

1. Use final module owner path `github.com/TrebuchetDynamics/gollmfree` consistently.
2. Confirm package layout: root package plus `providers/` and `cmd/gollmfree/`.
3. Create `go.mod` and baseline folders.
4. Add CI skeleton for `go test ./...`.

Exit criteria:

- Module path is final or explicitly marked temporary.
- Empty project builds or has only intentional TODO stubs.

### Milestone 1: Core API and Test Harness

**Goal:** Establish stable public types and fake-provider tests before real provider work.

Tasks:

1. Add `types.go` with request/response and stream chunk types.
2. Add `provider.go` with the provider interface.
3. Add `client.go` with `NewClient`, options, request validation, and fake-provider wiring for tests.
4. Add `registry.go` with concurrent-read-safe metadata lookup.
5. Add fake providers used only by tests.

Exit criteria:

- Unit tests prove client request validation and registry lookup.
- Public API examples compile against fake providers.

### Milestone 2: First Provider Vertical Slice

**Goal:** Prove one complete path from client to provider response.

Tasks:

1. Re-validate the current DeepAI endpoint and request format.
2. Implement `providers/deepai.go`.
3. Add mocked `httptest` coverage for DeepAI requests, parsing, provider errors, and streaming emulation.
4. Register DeepAI with aliases and default priority.
5. Add an opt-in real-provider smoke test.

Exit criteria:

- `Client.ChatCompletion` can return a DeepAI response.
- `go test ./...` passes without network.
- Real-provider smoke test is available but skipped by default.

### Milestone 3: Selector, Health, and Fallback

**Goal:** Make provider selection reliable before adding many providers.

Tasks:

1. Implement model candidate lookup and alias handling.
2. Implement `HealthStore` with mutex-protected snapshots.
3. Implement ranking, cooldown, sequential fallback, and combined errors.
4. Add `WithTimeout`, `WithMaxRetries`, `WithProviderPriority`, and `WithRaceMode`.
5. Add deterministic fake-provider tests for success, failure, fallback, cooldown, latency ordering, race mode, and concurrent client use.

Exit criteria:

- Tests prove fallback after provider failure.
- Tests prove repeated failures change ranking/cooldown.
- Tests prove health updates are concurrency-safe.

### Milestone 4: Provider Portfolio

**Goal:** Add enough providers for fallback to matter.

Tasks:

1. Re-check each candidate provider from `g4f` for current endpoint viability.
2. Implement ChatgptAi if endpoint remains simple and unauthenticated.
3. Implement Yqcloud if endpoint remains active.
4. Implement ChatgptLogin only if complexity stays low.
5. Stub or document Ails/You.com as postponed if session flow is too complex.
6. Add mocked tests and opt-in integration tests per included provider.

Exit criteria:

- At least three providers are implemented or explicitly documented with active/inactive/postponed status.
- Selector tests cover multiple registered providers.

### Milestone 5: CLI

**Goal:** Provide an installable smoke-test and user-facing command.

Tasks:

1. Implement `cmd/gollmfree/main.go` using the standard `flag` package.
2. Add `chat`, `chat --stream`, `list`, and `models` commands.
3. Ensure CLI uses `Client` and registry APIs only.
4. Add command smoke tests or scripted validation.

Exit criteria:

- `go install ./cmd/gollmfree` succeeds.
- `gollmfree models` and `gollmfree list` work without live provider calls.
- `chat` and `chat --stream` exercise the same library path as Go consumers.

### Milestone 6: `gormes-agent` Integration

**Goal:** Make the primary consumer path concrete.

Tasks:

1. Inspect the actual `gormes-agent` LLM interface.
2. Decide whether direct `Client` use is enough or an adapter is needed.
3. Add `examples/gormes-agent/` with a compiling example when possible.
4. Document limitations and expected setup.

Exit criteria:

- The example compiles, or the README records why the external dependency was unavailable.
- Provider selection remains hidden from the agent integration.

### Milestone 7: Release Hardening

**Goal:** Finish v0.1.0 documentation, quality gates, and release readiness.

Tasks:

1. Add GoDoc comments for all public types and methods.
2. Write README quick start, CLI usage, Go API usage, provider status, privacy caveats, and install steps.
3. Add examples for basic, streaming, CLI, and `gormes-agent`.
4. Run `go test ./...` and `go vet ./...`.
5. Tag `v0.1.0` after all acceptance criteria pass.

Exit criteria:

- README and examples cover all intended users.
- Local validation and CI pass.
- v0.1.0 definition of done is satisfied.

### 9.3 Detailed Backlog

| ID | Title | Scope | Depends on | Done signal |
| --- | --- | --- | --- | --- |
| T0.1 | Choose module path | `go.mod`, PRD, README placeholders | None | Module path is `github.com/TrebuchetDynamics/gollmfree`; no old placeholder path remains unless intentionally documented. |
| T0.2 | Scaffold module | `go.mod`, folders, `.gitignore` if needed | T0.1 | `go test ./...` runs successfully on scaffold. |
| T0.3 | Add CI skeleton | `.github/workflows/test.yml` | T0.2 | Workflow runs `go test ./...` on push/PR. |
| T0.4 | Create living README skeleton | `README.md` | T0.1 | README has project status, install/module path placeholder, quick start placeholders, provider status table, testing policy, and privacy caveats. |
| T1.1 | Public types | `types.go`, `README.md` | T0.2 | Types match FR-4, have GoDoc comments, and README quick-start API snippet is updated or explicitly marked pending. |
| T1.2 | Provider interface | `provider.go` | T1.1 | Interface matches FR-1 and fake provider compiles. |
| T1.3 | Registry | `registry.go`, `registry_test.go` | T1.2 | Tests cover aliases, duplicate names, unknown model, returned-copy behavior. |
| T1.4 | Client options | `options.go`, `client.go`, tests | T1.1 | Tests cover defaults and invalid option handling. |
| T1.5 | Fake-provider harness | `internal/testprovider` or test files | T1.2 | Selector/client tests can simulate success, failure, delay, stream. |
| T2.1 | DeepAI endpoint research | notes in provider comments/README | T1.2 | Current endpoint status is documented with date/source. |
| T2.2 | DeepAI implementation | `providers/deepai.go` | T2.1 | `httptest` happy path returns `CompletionResponse`. |
| T2.3 | DeepAI error tests | `providers/deepai_test.go` | T2.2 | Tests cover non-2xx, malformed JSON/body, context cancellation. |
| T2.4 | DeepAI stream emulation | `providers/deepai.go`, tests | T2.2 | Stream closes and emits one chunk for non-streaming response. |
| T2.5 | Integration smoke gate | `providers/deepai_integration_test.go` | T2.2 | Test is opt-in and skipped in normal `go test ./...`. |
| T3.1 | Health store | `health.go`, `health_test.go` | T1.5 | Race detector/concurrency test passes. |
| T3.2 | Ranking | `selector.go`, `selector_test.go` | T1.3, T3.1 | Tests prove priority, cooldown, failures, latency, deterministic tie-break. |
| T3.3 | Sequential fallback | `client.go`, `selector.go`, tests | T3.2 | Test proves failed provider falls through to next and combined error includes all failures. |
| T3.4 | Timeouts/retries | `client.go`, `options.go`, tests | T3.3 | Tests prove per-attempt timeout and bounded retry count. |
| T3.5 | Race mode | `selector.go`, `client.go`, tests | T3.3 | Deterministic fake test proves first success wins and canceled losers are not marked failed. |
| T4.1 | Provider viability pass | README provider table | T2.1 | ChatgptAi/Yqcloud/ChatgptLogin/Ails/You.com statuses recorded. |
| T4.2 | ChatgptAi provider | `providers/chatgptai.go`, tests | T4.1 | Mocked tests pass or provider is explicitly postponed. |
| T4.3 | Yqcloud provider | `providers/yqcloud.go`, tests | T4.1 | Mocked tests pass or provider is explicitly postponed. |
| T4.4 | Additional provider decision | provider file or docs | T4.1 | At least one of ChatgptLogin/Ails/You.com implemented or postponed with reason. |
| T5.1 | CLI command router | `cmd/gollmfree/main.go` | T1.4 | Usage errors return exit code 2. |
| T5.2 | CLI `models` | CLI + tests | T1.3 | Runs without network and prints aliases/providers. |
| T5.3 | CLI `list` | CLI + tests | T3.1 | Runs without network and prints provider health/status. |
| T5.4 | CLI `chat` | CLI + tests | T3.3 | Uses `Client.ChatCompletion`, supports `--model` and `--timeout`. |
| T5.5 | CLI stream | CLI + tests | T2.4 or native stream provider | Uses `Client.ChatCompletionStream`, flushes chunks. |
| T5.6 | README CLI refresh | `README.md` | T5.1-T5.5 | README CLI commands/flags match implemented behavior and no-network commands are called out. |
| T6.1 | Inspect `gormes-agent` | notes, README | Stable client API | Actual LLM interface is linked or unavailable status documented. |
| T6.2 | Adapter/example | `examples/gormes-agent/` | T6.1 | Example compiles or limitation is documented. |
| T7.1 | README final audit | `README.md` | T5 | Install, Go API, CLI, provider status, caveats, examples, validation commands, and maturity status are accurate against current code. |
| T7.2 | Examples | `examples/basic`, `examples/streaming` | T3/T5 | Examples compile or are tested by CI. |
| T7.3 | Quality gate | code/tests/CI | All | `go test ./...` and `go vet ./...` pass. |
| T7.4 | Release checklist | README/PRD tag notes | T7.3 | v0.1.0 DoD checklist is fully checked. |

### 9.4 Acceptance Checklist by Requirement

| Requirement area | Evidence to collect before release |
| --- | --- |
| FR-1 Provider Interface | `provider.go` signature, provider tests proving context cancellation and stream close. |
| FR-2 Registry | Registry tests for provider metadata, aliases, default priority, concurrent reads/copy returns. |
| FR-3 Selector | Unit tests for model aliases, priority override, dynamic health, cooldown, latency, fallback, combined errors, race mode. |
| FR-4 Client API | Compileable examples and tests for `NewClient`, sync chat, stream chat, options. |
| FR-5 CLI | `go install ./cmd/gollmfree`, no-network tests for `list`/`models`, fake-client tests for `chat`. |
| FR-6 `gormes-agent` | Inspected interface plus compiling adapter/example or documented external-repo blocker. |
| FR-7 No server/GUI/Docker | Dependency review and README statement; no server packages or Docker files required. |
| NFR concurrency | `go test -race ./...` for health/client tests where practical. |
| NFR resilience | Tests for deadlines, retries, cooldown expiry, combined errors. |
| Privacy/security | README caveat, no prompt logging by default, no credentials in provider APIs. |
| Living documentation | README exists from scaffolding and has been refreshed at each milestone that changed user-facing behavior/status. |

### 9.5 Known Gaps to Close During Implementation

- Final module owner path is resolved as `github.com/TrebuchetDynamics/gollmfree`; keep install docs/import examples aligned with it.
- Live status of every proposed provider must be re-verified; the PRD intentionally does not assume endpoints still work.
- `gormes-agent` interface is unknown until inspected.
- Stream error reporting after a stream starts is limited in v0.1.0; document this limitation if not solved.
- Race mode needs a policy for whether to continue sequentially after race candidates fail; recommended behavior is continue sequentially for v0.1.0.
- Provider status documentation must be updated at release time because anonymous endpoints change quickly.

## 10. Build Plan

### Phase 1: Foundation

**Goal:** Core interfaces, project scaffold, and one working provider.

Tasks:

1. Create Go module `github.com/TrebuchetDynamics/gollmfree`.
2. Define `Provider` in `provider.go`.
3. Define `Message`, `ChatRequest`, `CompletionResponse`, `Choice`, and stream chunk types.
4. Implement the simplest provider, initially DeepAI if still available.
5. Use mocked tests for provider request construction and response parsing.
6. Add an optional real-provider integration test skipped by default.
7. Create provider registry.

Deliverable:

- A Go package that can call one provider through the common provider interface.

Acceptance criteria:

- `go test ./...` passes.
- `go test -run TestDeepAI -tags=integration ./...` can be used for a real smoke test when network access is available.

### Phase 2: Provider Porting Sprint

**Goal:** Add a portfolio of providers.

Tasks for each provider:

1. Study current `g4f` implementation and verify endpoint still exists.
2. Port request headers, payload, response parsing, and errors to Go.
3. Put implementation in its own provider file.
4. Register provider name, default priority, and model aliases.
5. Add mocked tests.
6. Add skipped integration test where feasible.

Target providers:

- DeepAI
- ChatgptAi
- Yqcloud
- ChatgptLogin
- Ails, if stable
- You.com, postponed unless simple enough

Deliverable:

- Multiple registered providers covering common model aliases.

Acceptance criteria:

- `gollmfree models` shows model/provider coverage.
- Skipped integration tests exist for unstable external calls.
- Provider failures do not crash selector calls.

### Phase 3: Ranked Selector and Health Tracking

**Goal:** Automatic provider choice with fallback.

Tasks:

1. Implement `Selector` candidate lookup by model.
2. Implement health scoring with success count, failure count, consecutive failures, latency, last error, cooldown-until timestamp, and last success.
3. Sort candidates by user priority, default priority, cooldown, failure state, and latency.
4. Implement sequential fallback.
5. Implement optional race mode.
6. Implement combined errors.
7. Add options:
   - `WithTimeout(d time.Duration)`
   - `WithMaxRetries(n int)`
   - `WithProviderPriority(model string, priorities []string)`
   - `WithRaceMode(enabled bool)`
8. Add concurrency tests.

Deliverable:

- `Client.ChatCompletion` returns from the first healthy working provider and falls back automatically.

Acceptance criteria:

- Unit tests prove fallback after failure.
- Unit tests prove cooldown/de-ranking after repeated failures.
- Unit tests prove successful provider health improves.
- Race mode has a deterministic mock-provider test.

### Phase 4: CLI Implementation

**Goal:** Usable terminal command.

Tasks:

1. Create `cmd/gollmfree/main.go`.
2. Implement `chat`, `chat --stream`, `list`, and `models`.
3. Wire commands to `gollmfree.Client` and registry.
4. Add basic command tests or scripted smoke tests.

Deliverable:

- Installable `gollmfree` binary.

Acceptance criteria:

- `go install ./cmd/gollmfree` succeeds.
- `gollmfree models` prints known aliases.
- `gollmfree list` prints providers and status.

### Phase 5: `gormes-agent` Integration

**Goal:** Document and test drop-in agent integration.

Tasks:

1. Inspect `gormes-agent` LLM interface.
2. Add adapter if direct `Client` use is not enough.
3. Add `examples/gormes-agent/` with minimal setup.
4. Document zero-configuration behavior and any limitations.

Deliverable:

- Documented integration with `gormes-agent`.

Acceptance criteria:

- Example compiles, or README clearly states the expected interface if the dependent repository is unavailable.
- Provider selection remains hidden from the agent caller.

### Phase 6: Polish, Documentation, and Release

Tasks:

1. Add GoDoc comments for public types and methods.
2. Write README quick start.
3. Add provider list and caveats.
4. Add examples for basic, streaming, CLI, and `gormes-agent` usage.
5. Add GitHub Actions CI for `go test ./...`.
6. Run `go test ./...` and `go vet ./...`.
7. Tag `v0.1.0` after acceptance criteria pass.

Deliverable:

- A documented and testable v0.1.0-ready module.

## 11. TDD and Testing Strategy

All development must follow TDD unless the task is explicitly labeled as research, documentation, or a spike. A spike may inspect provider behavior or external interfaces, but any production code that survives the spike must be covered by tests before it is considered complete.

### 11.1 TDD Operating Rules

1. **Red first:** write a failing unit, integration, CLI, or compile/example test that describes the next behavior.
2. **Smallest green:** implement only enough production code to pass the focused test.
3. **Refactor after green:** improve locality, naming, and duplication only while tests stay passing.
4. **One behavior per slice:** avoid bundling registry, selector, provider, and CLI behavior in one red-green loop.
5. **No live-provider dependency in normal tests:** use fake providers and `httptest` for CI-safe behavior.
6. **Record evidence:** every completed backlog item needs a passing command or documented blocker.
7. **Regression tests for bugs:** if a bug is found, first add a failing test that reproduces it, then fix.

### 11.2 Test Pyramid

| Layer | Purpose | Tools | Must be deterministic? |
| --- | --- | --- | --- |
| Pure unit tests | Types, validation, registry, selector ranking, health scoring, error formatting. | `testing`, fake providers. | Yes. |
| Concurrency tests | Client and health safety under goroutines. | `testing`, `sync`, optional `go test -race`. | Yes. |
| Provider mocked tests | HTTP request construction and response parsing. | `httptest.Server`. | Yes. |
| CLI tests | Argument parsing, exit codes, no-network `list`/`models`, fake-client `chat`. | `testing`, command helpers. | Yes. |
| Compile/example tests | Public API examples remain accurate. | `go test ./examples/...` or package examples. | Yes. |
| Live integration tests | Smoke-test anonymous providers. | `//go:build integration`, env vars. | No; opt-in only. |

### 11.3 Required Red-Green Slices by Module

| Module | First failing test to write | Green implementation target | Refactor check |
| --- | --- | --- | --- |
| Types | Example compile test referencing all public fields. | Add `types.go`. | GoDoc and JSON tags match FR-4. |
| Provider interface | Fake provider compile assertion. | Add `provider.go`. | Interface stays narrow; no concrete provider leakage. |
| Registry | Duplicate provider name and alias lookup tests. | Immutable registry with normalized aliases. | Returned slices are defensive copies. |
| Options | Invalid timeout/retry tests. | Functional options with validated defaults. | No global mutable config. |
| Health | Success/failure/cooldown tests. | Mutex-protected `HealthStore`. | Snapshot calculation isolated from selector. |
| Selector | Ranking order table test. | Deterministic sort and candidate filtering. | Ranking has stable tie-breaks. |
| Client fallback | Fake first provider fails, second succeeds. | Attempt loop records health and returns second response. | Attempt orchestration readable and provider-agnostic. |
| Race mode | Faster successful fake wins; canceled loser not failed. | Race implementation with cancellation policy. | No goroutine leaks. |
| Provider | `httptest` happy path and malformed response tests. | Provider request/parse implementation. | Endpoint quirks stay in provider file. |
| CLI | Usage error exit code test. | Command router and handlers. | Handlers call library APIs, not providers directly. |

### 11.4 Validation Commands

Run the narrowest relevant command during red-green, then broader gates before marking a task done.

| Situation | Command |
| --- | --- |
| Focused package test | `go test ./... -run TestName` or `go test ./providers -run TestDeepAI` |
| Full unit suite | `go test ./...` |
| Race-sensitive changes | `go test -race ./...` |
| Static sanity | `go vet ./...` |
| CLI install gate | `go install ./cmd/gollmfree` |
| Opt-in live provider smoke | `go test -tags=integration ./... -run TestDeepAI` |

### 11.5 Evidence Receipt Format

When a task is completed, record evidence in the relevant tracker or issue using this shape:

```text
Task: T3.3 Sequential fallback
Red: TestClientFallsBackAfterProviderFailure failed with expected missing behavior
Green: go test ./... -run TestClientFallsBackAfterProviderFailure
Gate: go test ./...
Notes: CombinedError includes deepai and yqcloud failures in attempt order
```

### 11.6 Unit Tests

- Provider request generation against mock HTTP servers.
- Provider response parsing.
- Selector ordering.
- Fallback behavior.
- Cooldown behavior.
- Race mode.
- Client concurrent calls.
- CLI argument parsing.

### 11.7 Integration Tests

- Real provider tests are useful but unstable.
- Mark real-provider tests with an integration tag or skip in `testing.Short()`.
- Do not make CI depend on third-party anonymous provider uptime.

### 11.8 Manual Smoke Tests

```bash
go test ./...
go install ./cmd/gollmfree
gollmfree models
gollmfree list
gollmfree chat "Say hello in one sentence"
gollmfree chat --stream "Say hello in one sentence"
```

## 12. Risks and Mitigations

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Anonymous provider endpoints change frequently | High | Keep providers isolated; add health tracking; document instability. |
| Providers block scraping headers | High | Use multiple providers and cooldown fallback. |
| Model labels are misleading | Medium | Treat model names as claimed capabilities, expose provider in response. |
| Integration tests are flaky | Medium | Mock unit tests in CI; real tests opt-in. |
| Legal/terms ambiguity | Medium | Document that users are responsible for provider terms and acceptable use. |
| Race mode increases external traffic | Low/Medium | Make race mode configurable and disabled by default. |

## 13. Security, Privacy, and Compliance Notes

- Prompts are sent to third-party anonymous providers; users must not send secrets or sensitive data.
- The README MUST disclose that providers are not controlled by this project.
- The project MUST not collect credentials.
- Provider implementations SHOULD avoid logging prompt contents by default.
- Users are responsible for complying with upstream provider terms.

## 14. Success Metrics

For v0.1.0:

- At least one provider works through the full client path.
- At least three providers are implemented or stubbed behind reliable tests, with active/inactive status documented.
- `Client.ChatCompletion` supports fallback and combined errors.
- `Client.ChatCompletionStream` works with native or emulated streaming.
- CLI `chat`, `chat --stream`, `list`, and `models` are available.
- `go test ./...` passes locally and in CI.
- A `gormes-agent` integration example exists.

For later versions:

- Five or more live providers are available.
- Health scoring materially improves successful completion rate.
- Documentation clearly tracks provider status and caveats.

## 15. Open Decisions

1. Final module owner path is decided: `github.com/TrebuchetDynamics/gollmfree`.
2. Should initial providers live in package root or `providers/`? Recommended: `providers/` for maintainability.
3. Should race mode be enabled by default? Recommended: no, to reduce external traffic.
4. Should the CLI use only `flag` or adopt Cobra? Recommended: standard library `flag` for v0.1.0.
5. How exactly does `gormes-agent` model its LLM interface? Inspect before Phase 5.
6. Which `g4f` providers are currently live enough to include beyond DeepAI?

## 16. Initial Implementation Checklist

Update this checklist as work lands. Every checked item must have TDD evidence, except explicitly documentation-only items.

- [x] Choose final module path and update this PRD.
- [x] Create `go.mod`.
- [x] Create living `README.md` skeleton during scaffolding.
- [ ] Add `types.go`.
- [ ] Add `provider.go`.
- [ ] Add `registry.go`.
- [ ] Add first provider implementation.
- [ ] Add mock provider tests.
- [ ] Add TDD evidence receipts for each completed backlog item.
- [ ] Add `selector.go` and health tracking.
- [ ] Add `client.go`.
- [ ] Add CLI under `cmd/gollmfree`.
- [ ] Add examples.
- [ ] Maintain README after each milestone/user-facing behavior change.
- [ ] Add final README audit before release.
- [x] Add CI.
- [ ] Inspect and document `gormes-agent` integration.

## 17. Definition of Done for v0.1.0

Gollmfree v0.1.0 is done when:

- This PRD's Master Progress Tracker, Detailed Backlog, Decision Log, Blocker Log, and Initial Implementation Checklist are current.
- Every production behavior was built through TDD, with evidence recorded or a documented spike exception.
- A Go consumer can import the module and call `gollmfree.NewClient()` with no required config.
- `ChatCompletion` and `ChatCompletionStream` are implemented.
- At least one real provider works, and multiple providers are registered or documented with current status.
- Ranked selector fallback is covered by tests.
- Health tracking is concurrency-safe and covered by tests.
- CLI commands `chat`, `chat --stream`, `list`, and `models` work.
- `gormes-agent` integration is documented with an example or adapter.
- `go test ./...` passes.
- README was created early, maintained periodically, and accurately documents provider instability, privacy caveats, install steps, CLI usage, Go API usage, provider status, examples, validation commands, and project maturity.
