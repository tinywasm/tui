package tui

// This file holds the ONE place where handler-kind detection order is written.
// Both devtui (local registration) and app (daemon serialization) classify via
// Classify, so the order below is the single authority — nothing re-derives it.
//
// Detection uses structural (anonymous) interfaces so this package never needs
// to import the named handler interfaces from devtui; the dependency arrow is
// one-way (devtui/app -> tui, never back).

// spec describes one handler kind: its wire Kind, whether it renders a footer
// field, and the predicate that recognizes it.
type spec struct {
	kind     Kind
	hasField bool
	detect   func(any) bool
}

// Structural interfaces — the minimal method set that identifies each kind.
type (
	iName        interface{ Name() string }
	iContent     interface{ Content() string }
	iChange      interface{ Change(string) }
	iExecute     interface{ Execute() }
	iWaiting     interface{ WaitingForUser() bool }
	iOptions     interface{ Options() []map[string]string }
	iSetLog      interface{ SetLog(func(...any)) }
	iLabel       interface{ Label() string }
	iValue       interface{ Value() string }
	iShortcuts   interface{ Shortcuts() []map[string]string }
	iChangeValue interface {
		iName
		iLabel
		iValue
		iChange
	}
)

// specs is ORDERED by detection precedence (most specific first). The order is
// load-bearing: Interactive and Selection are supersets of Edit, so they must
// precede it; Loggable is the most general (almost every handler has Name+SetLog),
// so it is the catch-all last. Do not reorder without updating the reasoning.
var specs = []spec{
	{KindDisplay, true, func(h any) bool {
		_, hasName := h.(iName)
		_, hasContent := h.(iContent)
		return hasName && hasContent
	}},
	{KindInteractive, true, func(h any) bool {
		_, ok := h.(iChangeValue)
		_, waiting := h.(iWaiting)
		return ok && waiting
	}},
	{KindSelection, true, func(h any) bool {
		_, ok := h.(iChangeValue)
		_, opts := h.(iOptions)
		return ok && opts
	}},
	{KindExecution, true, func(h any) bool {
		_, hasName := h.(iName)
		_, hasLabel := h.(iLabel)
		_, hasExec := h.(iExecute)
		return hasName && hasLabel && hasExec
	}},
	{KindEdit, true, func(h any) bool {
		_, ok := h.(iChangeValue)
		return ok
	}},
	{KindLoggable, false, func(h any) bool {
		_, hasName := h.(iName)
		_, hasLog := h.(iSetLog)
		return hasName && hasLog
	}},
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
// filled only when the handler exposes the matching method.
func Extract(h any) Meta {
	var m Meta
	if v, ok := h.(iName); ok {
		m.Name = v.Name()
	}
	if v, ok := h.(iLabel); ok {
		m.Label = v.Label()
	}
	if v, ok := h.(iValue); ok {
		m.Value = v.Value()
	}
	if v, ok := h.(iOptions); ok {
		m.Options = v.Options()
	}
	if v, ok := h.(iShortcuts); ok {
		m.Shortcuts = v.Shortcuts()
	}
	return m
}
