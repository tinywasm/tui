package tui

// This file holds the ONE place where handler-kind detection order is written.
// Both devtui (local registration) and app (daemon serialization) classify via
// Classify, so the order below is the single authority — nothing re-derives it.
//
// Classify switches on the SAME named interfaces a handler author implements
// (HandlerDisplay, HandlerSelection, ...; see interfaces.go) — never a private
// re-declaration of their method sets. That is the point of this package: one
// definition of each contract, used both for documentation and for detection,
// so they cannot drift apart.

// spec describes one handler kind: its wire Kind, whether it renders a footer
// field, and the predicate that recognizes it.
type spec struct {
	kind     Kind
	hasField bool
	detect   func(any) bool
}

// specs is ORDERED by detection precedence (most specific first). The order is
// load-bearing: HandlerInteractive and HandlerSelection are supersets of
// HandlerEdit (same Name/Label/Value/Change, plus one more method), so they
// must precede it or every Interactive/Selection handler would be misdetected
// as a plain Edit field. HandlerDisplay and HandlerExecution are disjoint from
// the others (different method shapes) so their position doesn't matter, but
// they're kept ahead of Edit to mirror devtui's historical switch order.
// Loggable is the most general (only Name+SetLog) so it is the catch-all last.
var specs = []spec{
	{KindDisplay, true, func(h any) bool { _, ok := h.(HandlerDisplay); return ok }},
	{KindInteractive, true, func(h any) bool { _, ok := h.(HandlerInteractive); return ok }},
	{KindSelection, true, func(h any) bool { _, ok := h.(HandlerSelection); return ok }},
	{KindExecution, true, func(h any) bool { _, ok := h.(HandlerExecution); return ok }},
	{KindEdit, true, func(h any) bool { _, ok := h.(HandlerEdit); return ok }},
	{KindLoggable, false, func(h any) bool { _, ok := h.(Loggable); return ok }},
}

// Classify runs the single ordered detection walk and returns the handler's
// Kind and whether it renders a footer field. hasField is false only for
// KindLoggable (log stream, no field). A handler matching nothing is treated as
// a log-only handler with no field.
func Classify(h any) (kind Kind, hasField bool) {
	for _, s := range specs {
		if s.detect(h) {
			return s.kind, s.hasField
		}
	}
	return KindLoggable, false
}

// Extract reads the metadata a StateEntry needs off any handler. Each field is
// filled only when the handler exposes the matching method. Options and
// Shortcuts are read via the same named interfaces (HandlerSelection,
// ShortcutProvider) that Classify uses, not ad-hoc method checks.
func Extract(h any) Meta {
	var m Meta
	if v, ok := h.(interface{ Name() string }); ok {
		m.Name = v.Name()
	}
	if v, ok := h.(interface{ Label() string }); ok {
		m.Label = v.Label()
	}
	if v, ok := h.(interface{ Value() string }); ok {
		m.Value = v.Value()
	}
	if v, ok := h.(HandlerSelection); ok {
		m.Options = v.Options()
	}
	if v, ok := h.(ShortcutProvider); ok {
		m.Shortcuts = v.Shortcuts()
	}
	return m
}
