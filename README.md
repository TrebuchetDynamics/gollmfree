# gollmfree

> Status: early scaffold / pre-v0.1.0. APIs, provider availability, and CLI behavior are still being built through strict TDD from [`GOLLMFREE-PRD.md`](GOLLMFREE-PRD.md).

`gollmfree` is a pure Go library and CLI for routing chat-completion requests to currently available anonymous/free LLM providers without API keys, sign-up, browser automation, Docker, a server process, or extra infrastructure.

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

Planned public API shape:

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

## Provider status

| Provider | Status | Notes |
| --- | --- | --- |
| DeepAI | planned | First vertical-slice provider; live endpoint must be re-validated. |
| ChatgptAi | planned | Include only if current endpoint remains simple and unauthenticated. |
| Yqcloud | planned | Include only if endpoint remains active. |
| ChatgptLogin | planned | Lower priority; include only if complexity stays low. |
| Ails | postponed candidate | Include only if stable and not too complex. |
| You.com | postponed candidate | Postpone if session/header flow is complex. |

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
