# Third-Party Notices

This project uses upstream projects as implementation references. Do not vendor or copy third-party source into this repository unless the corresponding license obligations are reviewed and this notice is updated.

## xtekky/gpt4free (`g4f`)

- Repository: <https://github.com/xtekky/gpt4free>
- Local reference cache: `.upstream/gpt4free` (ignored by git; not vendored)
- Upstream commit inspected: `798d8586b180cd8e6fc4b2b2a6a0c8a410de22ca`
- License file inspected: `LICENSE`
- License identified from upstream: GNU General Public License v3.0 only (`GPL-3.0`)
- Legal notice inspected: `LEGAL_NOTICE.md`

`gollmfree` is a partial Go port/reference implementation. Provider behavior, model aliases, request shaping, selector/fallback ideas, and compatibility notes may be studied from `g4f`, then reimplemented intentionally in Go with tests. Python source from `g4f` must not be vendored into this repository.

When a provider or selector behavior is ported, record in `GOLLMFREE-PRD.md`:

1. the upstream commit SHA and exact files inspected;
2. request URL/method/headers/payload and response parsing/streaming/error behavior;
3. what was ported, omitted, postponed, or redesigned for Go;
4. attribution and license-impact notes.

Users are responsible for complying with third-party provider terms and applicable law. This file is an engineering notice, not legal advice.
