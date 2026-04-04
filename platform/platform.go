package platform

import "github.com/huanfeng/wind-ui/core"

// OSType identifies the host operating system.
type OSType int

const (
	OSWindows OSType = iota
	OSMacOS
	OSLinux
)

// ThemeMode represents the system-wide color theme.
type ThemeMode int

const (
	ThemeLight ThemeMode = iota
	ThemeDark
)

// Screen describes a physical display attached to the system.
type Screen struct {
	Bounds   core.Rect
	WorkArea core.Rect
	DPI      float64
	Primary  bool
}

// DialogResult enumerates possible outcomes of a modal dialog.
type DialogResult int

const (
	DialogOK     DialogResult = iota
	DialogCancel
	DialogYes
	DialogNo
)

// MessageDialogOptions configures a system message dialog.
type MessageDialogOptions struct {
	Title   string
	Message string
	// Type: Info / Warning / Error / Question
}

// FileDialogOptions configures a file-open/save dialog.
type FileDialogOptions struct {
	Title   string
	Filters []string // e.g., "*.png", "*.jpg"
}

// InputType enumerates soft-keyboard / input-field modes.
type InputType int

const (
	InputTypeText     InputType = iota
	InputTypeNumber
	InputTypePassword
)

// Platform is the top-level abstraction over the host OS.
type Platform interface {
	OS() OSType
	CreateWindow(opts WindowOptions) (Window, error)
	RunMainLoop()
	PostToMainThread(fn func())
	Quit()
	GetClipboard() Clipboard
	GetScreens() []Screen
	GetPrimaryScreen() Screen
	GetSystemLocale() string
	GetSystemTheme() ThemeMode
	CreateTextRenderer() core.TextRenderer
	CreateNativeEditText(parent Window) NativeEditText
	ShowMessageDialog(opts MessageDialogOptions) DialogResult
	ShowFileDialog(opts FileDialogOptions) (string, error)
	CreateTrayIcon() TrayIcon
	CreateClipboardMonitor() ClipboardMonitor
	CreateHotkeyManager() HotkeyManager
}
