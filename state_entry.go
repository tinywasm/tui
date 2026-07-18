package tui

// StateEntry is the JSON wire format for a single handler registered in the
// daemon TUI. Produced by the daemon (github.com/tinywasm/app's HeadlessTUI),
// consumed by the client (github.com/tinywasm/devtui client mode).
//
// The JSON tags are the published contract: any producer must match them
// exactly, and existing tags must never change. Only additive changes (new
// fields) are allowed.
type StateEntry struct {
	TabTitle     string              `json:"tab_title"`
	HandlerName  string              `json:"handler_name"`
	HandlerColor string              `json:"handler_color"`
	HandlerType  int                 `json:"handler_type"` // a Kind value
	Label        string              `json:"label"`
	Value        string              `json:"value"`
	Shortcut     string              `json:"shortcut"`  // primary key = handler Name()
	Shortcuts    []map[string]string `json:"shortcuts"` // from a Shortcuts() provider
	Options      []map[string]string `json:"options"`   // Selection buttons: ordered {value:label}
}

// Meta is the data Extract pulls off a handler. Fields are populated only when
// the handler exposes the corresponding method; the rest stay zero.
type Meta struct {
	Name      string
	Label     string
	Value     string
	Options   []map[string]string
	Shortcuts []map[string]string
}
