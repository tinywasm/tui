// Package tui is the single source of truth for the TUI handler-kind contract
// shared by github.com/tinywasm/devtui (the renderer/consumer) and
// github.com/tinywasm/app (the daemon producer).
//
// A "handler" is any value a consumer registers with the TUI. The TUI inspects
// the handler's Go interface to decide how to render it (button, text field,
// radio group, log stream…) and, in client/daemon mode, how to serialize it
// over the wire (StateEntry). Historically each consumer re-derived that
// mapping independently, so adding or reordering a handler kind silently broke
// one side. This package centralizes the whole contract:
//
//   - Kind             — the frozen enum (also the wire value).
//   - Classify/Extract — the ONE ordered interface-detection walk + metadata.
//   - StateEntry       — the JSON wire format.
//   - AllKinds         — lets every consumer write a guardrail test that fails
//     loudly when a new Kind is not fully wired.
//
// It has zero heavy dependencies (no charmbracelet/bubbletea/lipgloss), so the
// daemon can import it without pulling the terminal renderer.
package tui
