package windows

import (
	"syscall"

	"gowui/core"
	"gowui/platform"
)

// WindowsPlatform implements platform.Platform for Windows.
type WindowsPlatform struct {
	quitCh chan struct{}
}

// NewPlatform creates a new Windows platform instance.
// Sets Per-Monitor V2 DPI awareness so windows render at full resolution.
func NewPlatform() *WindowsPlatform {
	enableDPIAwareness()
	return &WindowsPlatform{quitCh: make(chan struct{})}
}

// enableDPIAwareness sets the process as Per-Monitor V2 DPI-aware.
// This ensures GetClientRect returns physical pixels and text renders crisply.
func enableDPIAwareness() {
	user32 := syscall.NewLazyDLL("user32.dll")

	// Try Per-Monitor V2 first (Windows 10 1703+)
	setDpiAwarenessCtx := user32.NewProc("SetProcessDpiAwarenessContext")
	if setDpiAwarenessCtx.Find() == nil {
		// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
		setDpiAwarenessCtx.Call(^uintptr(3)) // -4 as uintptr
		return
	}

	// Fallback: Per-Monitor V1 (Windows 8.1+)
	shcore := syscall.NewLazyDLL("shcore.dll")
	setDpiAwareness := shcore.NewProc("SetProcessDpiAwareness")
	if setDpiAwareness.Find() == nil {
		setDpiAwareness.Call(2) // PROCESS_PER_MONITOR_DPI_AWARE
	}
}

func (p *WindowsPlatform) OS() platform.OSType { return platform.OSWindows }

func (p *WindowsPlatform) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	return newWin32Window(p, opts)
}

func (p *WindowsPlatform) RunMainLoop() {
	runMessageLoop(p.quitCh)
}

func (p *WindowsPlatform) PostToMainThread(fn func()) {
	// Phase 1: execute directly (single-threaded usage)
	fn()
}

func (p *WindowsPlatform) Quit() {
	close(p.quitCh)
	postQuitMessage()
}

func (p *WindowsPlatform) CreateTextRenderer() core.TextRenderer {
	return CreateTextRendererWithFallback()
}

// Stubs for Phase 1

func (p *WindowsPlatform) GetClipboard() platform.Clipboard {
	return &stubClipboard{}
}

func (p *WindowsPlatform) GetScreens() []platform.Screen {
	return []platform.Screen{p.GetPrimaryScreen()}
}

func (p *WindowsPlatform) GetPrimaryScreen() platform.Screen {
	return platform.Screen{DPI: 96, Primary: true, Bounds: core.Rect{Width: 1920, Height: 1080}}
}

func (p *WindowsPlatform) GetSystemLocale() string {
	return "en"
}

func (p *WindowsPlatform) GetSystemTheme() platform.ThemeMode {
	return platform.ThemeLight
}

func (p *WindowsPlatform) CreateNativeEditText(parent platform.Window) platform.NativeEditText {
	return nil
}

func (p *WindowsPlatform) ShowMessageDialog(opts platform.MessageDialogOptions) platform.DialogResult {
	return platform.DialogOK
}

func (p *WindowsPlatform) ShowFileDialog(opts platform.FileDialogOptions) (string, error) {
	return "", nil
}

// stubClipboard implements platform.Clipboard as a no-op for Phase 1.
type stubClipboard struct {
	text string
}

func (c *stubClipboard) GetText() (string, error) {
	return c.text, nil
}

func (c *stubClipboard) SetText(text string) error {
	c.text = text
	return nil
}

func (c *stubClipboard) HasText() bool {
	return c.text != ""
}
