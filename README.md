# tui

Single source of truth for the **TUI handler-kind contract** shared by
[`tinywasm/devtui`](https://github.com/tinywasm/devtui) (the terminal renderer)
and [`tinywasm/app`](https://github.com/tinywasm/app) (the daemon that
serializes handler state to the client).

## Why this exists

A "handler" is any value a consumer registers with the TUI. The TUI inspects
the handler's Go interface to decide how to render it and how to serialize it
over the daemon↔client wire. That mapping used to be re-derived independently in
several places (`devtui`'s local switch, its wire constants, `app`'s hand-copied
constants and a second detection switch). Adding or reordering a handler kind
silently broke one side — e.g. a `Selection` handler (radio buttons) collapsing
into a plain text field in `tinywasm -tui`.

This package centralizes the entire contract so there is exactly one place to
change, and a completeness test that fails loudly when a new kind isn't fully
wired.

## Surface

| Symbol | Purpose |
|---|---|
| `Kind` + `Kind*` consts | The frozen handler-kind enum (also the wire value). Append-only. |
| `AllKinds()` | Every kind in wire order — iterate it in consumer guardrail tests. |
| `Classify(h any) (Kind, hasField bool)` | The one ordered interface-detection walk. |
| `Extract(h any) Meta` | Pulls `Name/Label/Value/Options/Shortcuts` off a handler. |
| `StateEntry` | The JSON wire format (produced by the daemon, consumed by the client). |

## The guardrail

`AllKinds()` + `Classify` let every consumer write a test that iterates all
kinds and asserts it handles each one. Appending a `Kind` here turns those tests
red until detection, serialization, reconstruction and rendering are all wired —
instead of degrading silently. See `classify_test.go` for the in-repo
completeness checks.

## Zero heavy dependencies

Detection uses structural interfaces, so this package imports nothing from
`devtui` and no `charmbracelet`/`bubbletea`/`lipgloss`. The daemon can import it
without pulling the terminal renderer. The dependency arrow is one-way:
`devtui`/`app` → `tui`, never back.

## Adding a new handler kind

1. Append a `Kind<Name>` constant (new highest value — never renumber).
2. Add it to `AllKinds()`.
3. Add a `spec` (with its detection predicate, in precedence order) to `specs`.
4. Add a sample to `sampleFor` in the test.
5. `go test ./...` — the completeness tests tell you what's missing.
6. Bump the version; consumers wire the new kind, guided by their own
   `AllKinds()` guardrail tests going red.
