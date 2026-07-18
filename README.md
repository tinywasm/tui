# tui
<img src="docs/img/badges.svg">

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
| `HandlerDisplay`, `HandlerEdit`, `HandlerExecution`, `HandlerInteractive`, `HandlerSelection`, `Loggable`, `StreamingLoggable`, `ShortcutProvider`, `Cancelable`, `TabAware` | The named interfaces a handler implements. **The whole contract lives here — nowhere else.** |
| `Kind` + `Kind*` consts | The frozen handler-kind enum (also the wire value). Append-only. |
| `AllKinds()` | Every kind in wire order — iterate it in consumer guardrail tests. |
| `Classify(h any) (Kind, hasField bool)` | The one ordered walk, switching on the interfaces above directly — no shadow copies. |
| `Extract(h any) Meta` | Pulls `Name/Label/Value/Options/Shortcuts` off a handler. |
| `StateEntry` | The JSON wire format (produced by the daemon, consumed by the client). |

## Why `interfaces.go` lives here, not in `devtui`

Originally `devtui` owned the named interfaces (`HandlerSelection`, etc.) as
documentation for handler authors, while its detection switch and the daemon's
detection switch each re-declared their own anonymous copy of the same method
shapes to do the actual type-switching. Three definitions of "what a Selection
handler looks like" — the exact drift this package exists to prevent, just one
level deeper. `Classify` now type-switches on `HandlerSelection` etc.
*directly*, so there is one shape, used for both the doc a handler author reads
and the detection that renders/serializes it. If you edit an interface here,
`Classify` sees the change immediately — there is nothing else to keep in sync.

## The guardrail

`AllKinds()` + `Classify` let every consumer write a test that iterates all
kinds and asserts it handles each one. Appending a `Kind` here turns those tests
red until the interface, detection, serialization, reconstruction and rendering
are all wired — instead of degrading silently. See `classify_test.go` for the
in-repo completeness checks.

## Zero heavy dependencies

Detection uses structural interfaces, so this package imports nothing from
`devtui` and no `charmbracelet`/`bubbletea`/`lipgloss`. The daemon can import it
without pulling the terminal renderer. The dependency arrow is one-way:
`devtui`/`app` → `tui`, never back.

## Adding a new handler kind

1. Add the named interface to `interfaces.go` (e.g. `HandlerFoo`).
2. Append a `Kind<Name>` constant (new highest value — never renumber).
3. Add it to `AllKinds()`.
4. Add a `spec` to `specs` in `classify.go`, switching on the interface from
   step 1 — never redeclare its method set inline.
5. Add a sample to `sampleFor` in the test.
6. `go test ./...` — the completeness tests tell you what's missing.
7. Bump the version; consumers wire the new kind, guided by their own
   `AllKinds()` guardrail tests going red.
