package windows

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/huanfeng/go-wui/platform"
)

// Win32 clipboard constants
const (
	CF_TEXT        = 1
	CF_UNICODETEXT = 13
	CF_HTML        = 0 // registered format, resolved at runtime

	WM_CLIPBOARDUPDATE = 0x031D

	GMEM_MOVEABLE = 0x0002
)

var (
	procAddClipboardFormatListener    = user32.NewProc("AddClipboardFormatListener")
	procRemoveClipboardFormatListener = user32.NewProc("RemoveClipboardFormatListener")
	procOpenClipboard                 = user32.NewProc("OpenClipboard")
	procCloseClipboard                = user32.NewProc("CloseClipboard")
	procGetClipboardData              = user32.NewProc("GetClipboardData")
	procIsClipboardFormatAvailable    = user32.NewProc("IsClipboardFormatAvailable")
	procSetClipboardData              = user32.NewProc("SetClipboardData")
	procEmptyClipboard                = user32.NewProc("EmptyClipboard")

	procGlobalLock   = kernel32.NewProc("GlobalLock")
	procGlobalUnlock = kernel32.NewProc("GlobalUnlock")
	procGlobalAlloc  = kernel32.NewProc("GlobalAlloc")
	procGlobalSize   = kernel32.NewProc("GlobalSize")
	procLstrcpyW     = kernel32.NewProc("lstrcpyW")

	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
)

// win32ClipboardMonitor implements platform.ClipboardMonitor using Win32 APIs.
type win32ClipboardMonitor struct {
	onChange func(content platform.ClipboardContent)
	started  bool
	lastText string // deduplicate consecutive identical copies
}

// NewClipboardMonitor creates a new clipboard monitor for Windows.
func NewClipboardMonitor() platform.ClipboardMonitor {
	return &win32ClipboardMonitor{}
}

func (cm *win32ClipboardMonitor) SetOnClipboardChanged(fn func(content platform.ClipboardContent)) {
	cm.onChange = fn
}

func (cm *win32ClipboardMonitor) Start() error {
	if cm.started {
		return nil
	}

	mw := getMessageWindow()

	// Register clipboard format listener
	ret, _, err := procAddClipboardFormatListener.Call(mw.Handle())
	if ret == 0 {
		return fmt.Errorf("AddClipboardFormatListener failed: %w", err)
	}

	// Install handler on the message window
	mw.clipboardHandler = cm.handleMessage

	cm.started = true
	return nil
}

func (cm *win32ClipboardMonitor) Stop() {
	if !cm.started {
		return
	}

	mw := getMessageWindow()
	procRemoveClipboardFormatListener.Call(mw.Handle())
	mw.clipboardHandler = nil
	cm.started = false
}

func (cm *win32ClipboardMonitor) handleMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) {
	if cm.onChange == nil {
		return
	}

	content := platform.ClipboardContent{
		Timestamp: time.Now(),
	}

	// Get foreground window title as source
	content.Source = getForegroundWindowTitle()

	// Read clipboard text
	text, ok := readClipboardText()
	if ok && text != "" {
		// Deduplicate
		if text == cm.lastText {
			return
		}
		cm.lastText = text
		content.Text = text
	}

	// Check for image
	ret, _, _ := procIsClipboardFormatAvailable.Call(uintptr(2)) // CF_BITMAP
	content.HasImage = ret != 0

	// Only fire callback if we have content
	if content.Text != "" || content.HasImage {
		cm.onChange(content)
	}
}

// readClipboardText reads Unicode text from the clipboard.
func readClipboardText() (string, bool) {
	ret, _, _ := procIsClipboardFormatAvailable.Call(CF_UNICODETEXT)
	if ret == 0 {
		return "", false
	}

	ret, _, _ = procOpenClipboard.Call(0)
	if ret == 0 {
		return "", false
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if h == 0 {
		return "", false
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return "", false
	}
	defer procGlobalUnlock.Call(h)

	// Read UTF-16 string
	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(ptr))[:])
	return text, true
}

// writeClipboardText writes Unicode text to the clipboard.
func writeClipboardText(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	ret, _, e := procOpenClipboard.Call(0)
	if ret == 0 {
		return e
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	size := len(utf16) * 2 // uint16 = 2 bytes
	hMem, _, e := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if hMem == 0 {
		return e
	}

	ptr, _, e := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return e
	}

	// Copy UTF-16 data
	src := unsafe.Pointer(&utf16[0])
	dst := unsafe.Pointer(ptr)
	copy((*[1 << 20]uint16)(dst)[:len(utf16)], (*[1 << 20]uint16)(src)[:len(utf16)])

	procGlobalUnlock.Call(hMem)
	procSetClipboardData.Call(CF_UNICODETEXT, hMem)

	return nil
}

// getForegroundWindowTitle returns the title of the currently active window.
func getForegroundWindowTitle() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}

	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}

	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}
