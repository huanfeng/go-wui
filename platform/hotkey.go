package platform

// KeyModifier represents keyboard modifier flags for hotkeys.
type KeyModifier uint32

const (
	ModAlt   KeyModifier = 0x0001
	ModCtrl  KeyModifier = 0x0002
	ModShift KeyModifier = 0x0004
	ModWin   KeyModifier = 0x0008
)

// HotkeyManager registers and manages global hotkeys.
type HotkeyManager interface {
	// Register adds a global hotkey binding.
	// The id is a user-defined string for later unregistration.
	Register(id string, modifiers KeyModifier, key uint32, handler func()) error
	// Unregister removes a previously registered hotkey.
	Unregister(id string) error
	// UnregisterAll removes all registered hotkeys.
	UnregisterAll()
}
