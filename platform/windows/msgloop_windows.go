package windows

import (
	"runtime"
	"unsafe"
)

// getMessage wraps the Win32 GetMessageW function.
func getMessage(msg *MSG, hwnd uintptr, msgFilterMin, msgFilterMax uint32) (int32, error) {
	ret, _, err := procGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		hwnd,
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return int32(ret), err
}

// translateMessage wraps the Win32 TranslateMessage function.
func translateMessage(msg *MSG) {
	procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

// dispatchMessage wraps the Win32 DispatchMessageW function.
func dispatchMessage(msg *MSG) {
	procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

// postQuitMessage wraps the Win32 PostQuitMessage function.
func postQuitMessage() {
	procPostQuitMessage.Call(0)
}

// runMessageLoop runs the Win32 message loop until quit is signalled.
func runMessageLoop(quit chan struct{}) {
	runtime.LockOSThread()
	var msg MSG
	for {
		select {
		case <-quit:
			return
		default:
		}
		ret, _ := getMessage(&msg, 0, 0, 0)
		if ret == 0 {
			break // WM_QUIT received
		}
		if ret == -1 {
			break // Error
		}
		// IsDialogMessage handles Tab/Enter for child controls (EditText).
		// If it processes the message, skip normal dispatch.
		isDialog, _, _ := procIsDialogMessageW.Call(msg.HWnd, uintptr(unsafe.Pointer(&msg)))
		if isDialog != 0 {
			continue
		}
		translateMessage(&msg)
		dispatchMessage(&msg)
	}
}
