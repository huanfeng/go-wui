package platform

import "time"

// ClipboardContent represents a single clipboard change event.
type ClipboardContent struct {
	Text      string
	HTML      string // optional rich text content
	HasImage  bool
	Timestamp time.Time
	Source    string // foreground window title at the time of copy
}

// ClipboardMonitor watches the system clipboard for changes.
type ClipboardMonitor interface {
	// Start begins monitoring clipboard changes.
	Start() error
	// Stop ends clipboard monitoring and releases resources.
	Stop()
	// SetOnClipboardChanged registers a callback for clipboard change events.
	SetOnClipboardChanged(fn func(content ClipboardContent))
}
