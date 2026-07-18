package tui

// Kind identifies how a registered handler is rendered and serialized.
//
// The integer values are the FROZEN wire protocol: they are written into
// StateEntry.HandlerType and read back by remote consumers. Never reorder or
// renumber existing kinds — only append new ones at the end. TestKindWireValues
// guards this.
type Kind int

const (
	KindDisplay     Kind = 0 // read-only footer text (Name + Content)
	KindEdit        Kind = 1 // free-text input field (Name + Label + Value + Change)
	KindExecution   Kind = 2 // action button (Name + Label + Execute)
	KindInteractive Kind = 3 // auto-editing field (Edit + WaitingForUser)
	KindLoggable    Kind = 4 // log stream only, no footer field (Name + SetLog)
	KindSelection   Kind = 5 // radio / segmented buttons (Edit + Options)
)

// String returns the kind's name, for readable log and test output.
func (k Kind) String() string {
	switch k {
	case KindDisplay:
		return "Display"
	case KindEdit:
		return "Edit"
	case KindExecution:
		return "Execution"
	case KindInteractive:
		return "Interactive"
	case KindLoggable:
		return "Loggable"
	case KindSelection:
		return "Selection"
	default:
		return "Unknown"
	}
}

// AllKinds returns every declared Kind in wire order. Consumers iterate this in
// their own guardrail tests to assert they handle every kind — so a newly
// appended Kind turns those tests red instead of degrading silently.
func AllKinds() []Kind {
	return []Kind{
		KindDisplay,
		KindEdit,
		KindExecution,
		KindInteractive,
		KindLoggable,
		KindSelection,
	}
}
