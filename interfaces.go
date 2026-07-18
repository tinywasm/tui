package tui

// This file is the handler contract itself: the named interfaces a consumer
// implements to be recognized as a given Kind. It moved here from devtui so
// there is exactly ONE definition of "what a Selection handler looks like" —
// previously devtui named it here while a second, anonymous copy of the same
// method set lived in the detection switch (in this package and in app's
// daemon), and the two drifted independently. Classify (classify.go) type-
// switches on these interfaces directly; it no longer re-derives its own
// shape for each kind.
//
// All interfaces here use only stdlib types, so a handler author never needs
// to import this package to satisfy one — Go's structural typing is enough
// ("Zero Coupling"). Importing it is only useful for compile-time assertions
// or to read Kind/StateEntry.

// HandlerDisplay defines the interface for read-only information display handlers.
// These handlers show static or dynamic content without user interaction.
type HandlerDisplay interface {
	Name() string    // Full text to display in footer (handler responsible for content) eg. "System Status Information Display"
	Content() string // Display content (e.g., "help\n1-..\n2-...", "executing deploy wait...")
}

// HandlerEdit defines the interface for interactive fields that accept user input.
// These handlers allow users to modify values through text input.
type HandlerEdit interface {
	Name() string           // Identifier for logging: "ServerPort", "DatabaseURL"
	Label() string          // Field label (e.g., "Server Port", "Host Configuration")
	Value() string          // Current/initial value (e.g., "8080", "localhost")
	Change(newValue string) // Handle user input + content display via log
}

// HandlerExecution defines the interface for action buttons that execute operations.
// These handlers trigger business logic when activated by the user.
type HandlerExecution interface {
	Name() string  // Identifier for logging: "DeployProd", "BuildProject"
	Label() string // Button label (e.g., "Deploy to Production", "Build Project")
	Execute()      // Execute action + content display via log
}

// HandlerInteractive defines the interface for interactive content handlers.
// These handlers combine content display with user interaction capabilities.
// All content display is handled through progress() for consistency.
type HandlerInteractive interface {
	Name() string           // Identifier for logging: "ChatBot", "ConfigWizard"
	Label() string          // Field label (updates dynamically)
	Value() string          // Current input value
	Change(newValue string) // Handle user input + content display via log
	WaitingForUser() bool   // Should edit mode be auto-activated?
}

// HandlerSelection defines a radio / segmented-control field: a group of
// mutually-exclusive options rendered as evenly distributed buttons in the
// footer. Exactly one option is active at a time.
//
// It is a superset of HandlerEdit: Name/Label/Value/Change behave identically,
// and Options() advertises the discrete choice set that the TUI renders as
// buttons instead of a free-text input. Classify checks HandlerSelection
// BEFORE HandlerEdit for exactly this reason (see classify.go's specs order).
//
// Interaction: the user presses Enter to enter selection mode, moves the
// highlight with Left/Right, presses Enter to confirm (which calls
// Change(selectedValue)), or Esc to cancel.
//
// Options() returns an ORDERED slice of single-entry maps {value: label},
// identical in shape to ShortcutProvider.Shortcuts(). The map KEY is the
// stable value passed to Change(); the map VALUE is the human-readable button
// caption. Value() must return the key of the currently active option.
type HandlerSelection interface {
	Name() string                 // Identifier for logging: "CompilerMode"
	Label() string                // Group label shown before the buttons: "Compiler Mode"
	Value() string                // Key of the currently active option (e.g. "L")
	Change(newValue string)       // Called on confirm with the selected option's key
	Options() []map[string]string // Ordered {value: label} pairs (>= 2 recommended)
}

// ShortcutProvider defines the optional interface for handlers that provide global shortcuts.
// HandlerEdit/HandlerSelection implementations can implement this to enable global shortcut keys.
type ShortcutProvider interface {
	Shortcuts() []map[string]string // Returns ordered list of single-entry maps with shortcut->description, preserving registration order
}

// Cancelable defines the optional interface for handlers that want to be notified when the user cancels.
// Interactive handlers can implement this to clean up or reset their state when ESC is pressed.
type Cancelable interface {
	Cancel() // Called when user presses ESC to exit interactive mode
}

// TabAware defines the optional interface for handlers that want to be notified
// when their tab becomes active. This is useful for lazy initialization or
// delayed logging that requires the TUI's logger to be injected first.
type TabAware interface {
	OnTabActive()
}

// Loggable defines optional logging capability for handlers.
// Handlers implementing this receive a logger function from the TUI
// when registered via AddHandler.
//
// The log function provided by the TUI:
// - Is never nil (safe to call immediately)
// - Automatically tracks messages by handler Name()
// - Stores full history internally
// - Displays only most recent log in terminal (clean view)
//
// Example implementation:
//
//	type WasmClient struct {
//	    log func(message ...any)
//	}
//
//	func NewWasmClient() *WasmClient {
//	    return &WasmClient{
//	        log: func(message ...any) {}, // no-op until SetLog called
//	    }
//	}
//
//	func (w *WasmClient) Name() string { return "WASM" }
//
//	func (w *WasmClient) SetLog(logger func(message ...any)) {
//	    w.log = logger
//	}
//
//	func (w *WasmClient) Compile() {
//	    w.log("Compiling...")
//	}
type Loggable interface {
	Name() string
	SetLog(logger func(message ...any))
}

// StreamingLoggable enables handlers to display ALL log messages
// instead of the default "last message only" behavior.
type StreamingLoggable interface {
	Loggable
	AlwaysShowAllLogs() bool // Return true to show all messages
}

// LogOpen and LogClose are special prefixes for progress indication.
// Use LogOpen at the start of a long operation to show an animated spinner.
// Use LogClose when the operation completes to stop the animation.
//
// Example:
//
//	handler.log(tui.LogOpen, "Deploying to production")
//	// ... long operation ...
//	handler.log(tui.LogClose, "Deployment complete")
const (
	LogOpen  = "[..." // Start or update same line with auto-animation
	LogClose = "...]" // Update same line and stop auto-animation
)
