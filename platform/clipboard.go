package platform

// Clipboard provides read/write access to the system clipboard.
type Clipboard interface {
	GetText() (string, error)
	SetText(text string) error
	HasText() bool
}
