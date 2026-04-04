package windows

import (
	"fmt"
	"sync"

	"github.com/huanfeng/wind-ui/platform"
)

var (
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
)

// hotkeyEntry stores the mapping from numeric ID to user callback.
type hotkeyEntry struct {
	numID    int32
	handler  func()
}

// win32HotkeyManager implements platform.HotkeyManager using Win32 RegisterHotKey.
type win32HotkeyManager struct {
	mu       sync.Mutex
	hotkeys  map[string]*hotkeyEntry // string ID -> entry
	nextID   int32
}

// NewHotkeyManager creates a new global hotkey manager for Windows.
func NewHotkeyManager() platform.HotkeyManager {
	hm := &win32HotkeyManager{
		hotkeys: make(map[string]*hotkeyEntry),
		nextID:  1,
	}

	// Install WM_HOTKEY handler on the message window
	mw := getMessageWindow()
	mw.hotkeyHandler = hm.handleMessage

	return hm
}

func (hm *win32HotkeyManager) Register(id string, modifiers platform.KeyModifier, key uint32, handler func()) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Unregister if already exists
	if existing, ok := hm.hotkeys[id]; ok {
		mw := getMessageWindow()
		procUnregisterHotKey.Call(mw.Handle(), uintptr(existing.numID))
		delete(hm.hotkeys, id)
	}

	numID := hm.nextID
	hm.nextID++

	mw := getMessageWindow()
	ret, _, err := procRegisterHotKey.Call(
		mw.Handle(),
		uintptr(numID),
		uintptr(modifiers),
		uintptr(key),
	)
	if ret == 0 {
		return fmt.Errorf("RegisterHotKey failed for %q: %w", id, err)
	}

	hm.hotkeys[id] = &hotkeyEntry{
		numID:   numID,
		handler: handler,
	}

	return nil
}

func (hm *win32HotkeyManager) Unregister(id string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	entry, ok := hm.hotkeys[id]
	if !ok {
		return fmt.Errorf("hotkey %q not registered", id)
	}

	mw := getMessageWindow()
	ret, _, err := procUnregisterHotKey.Call(mw.Handle(), uintptr(entry.numID))
	if ret == 0 {
		return fmt.Errorf("UnregisterHotKey failed for %q: %w", id, err)
	}

	delete(hm.hotkeys, id)
	return nil
}

func (hm *win32HotkeyManager) UnregisterAll() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	mw := getMessageWindow()
	for _, entry := range hm.hotkeys {
		procUnregisterHotKey.Call(mw.Handle(), uintptr(entry.numID))
	}
	hm.hotkeys = make(map[string]*hotkeyEntry)
}

func (hm *win32HotkeyManager) handleMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) {
	hm.mu.Lock()
	numID := int32(wParam)
	var handler func()
	for _, entry := range hm.hotkeys {
		if entry.numID == numID {
			handler = entry.handler
			break
		}
	}
	hm.mu.Unlock()

	if handler != nil {
		handler()
	}
}
