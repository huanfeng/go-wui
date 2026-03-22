package windows

import (
	"gowui/core"
	"gowui/platform"
)

// WindowsPlatform implements platform.Platform for Windows.
type WindowsPlatform struct {
	quitCh chan struct{}
}

// NewPlatform creates a new Windows platform instance.
func NewPlatform() *WindowsPlatform {
	return &WindowsPlatform{quitCh: make(chan struct{})}
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
