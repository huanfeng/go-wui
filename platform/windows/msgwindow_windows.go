package windows

import (
	"sync"
	"syscall"
	"unsafe"
)

// Message-only hidden window for receiving system messages
// (clipboard updates, hotkeys, tray icon callbacks).
// Shared by ClipboardMonitor, TrayIcon, and HotkeyManager.

const (
	HWND_MESSAGE = ^uintptr(2) // (HWND)-3 = HWND_MESSAGE

	WM_HOTKEY = 0x0312

	// Custom message IDs for the message window
	WM_APP_TRAY      = WM_USER + 100
	WM_APP_CLIPBOARD = WM_USER + 101
)

var (
	msgWindow     *messageWindow
	msgWindowOnce sync.Once
)

// messageWindow is a hidden window that receives system messages.
type messageWindow struct {
	hwnd uintptr

	// Clipboard monitoring
	clipboardHandler func(hwnd uintptr, msg uint32, wParam, lParam uintptr)

	// Tray icon
	trayHandler func(hwnd uintptr, msg uint32, wParam, lParam uintptr)

	// Hotkey
	hotkeyHandler func(hwnd uintptr, msg uint32, wParam, lParam uintptr)
}

// getMessageWindow returns the singleton message-only window, creating it if needed.
func getMessageWindow() *messageWindow {
	msgWindowOnce.Do(func() {
		msgWindow = &messageWindow{}
		msgWindow.create()
	})
	return msgWindow
}

func (mw *messageWindow) create() {
	className, _ := syscall.UTF16PtrFromString("GoWUI_MsgWindow")

	hInst, _, _ := procGetModuleHandleW.Call(0)

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(msgWndProc),
		HInstance:     hInst,
		LpszClassName: className,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		0, // no title
		0, // no style
		0, 0, 0, 0,
		HWND_MESSAGE, // message-only window
		0, hInst, 0,
	)
	mw.hwnd = hwnd
}

// Handle returns the native window handle.
func (mw *messageWindow) Handle() uintptr {
	return mw.hwnd
}

// msgWndProc is the window procedure for the message-only window.
func msgWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	mw := msgWindow
	if mw == nil {
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}

	switch {
	case msg == WM_CLIPBOARDUPDATE && mw.clipboardHandler != nil:
		mw.clipboardHandler(hwnd, msg, wParam, lParam)
		return 0

	case msg == WM_APP_TRAY && mw.trayHandler != nil:
		mw.trayHandler(hwnd, msg, wParam, lParam)
		return 0

	case msg == WM_HOTKEY && mw.hotkeyHandler != nil:
		mw.hotkeyHandler(hwnd, msg, wParam, lParam)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}
