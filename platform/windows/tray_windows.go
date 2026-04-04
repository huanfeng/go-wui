package windows

import (
	"syscall"
	"unsafe"

	"github.com/huanfeng/wind-ui/platform"
)

// Win32 tray/menu constants
const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_INFO    = 0x00000010

	NIIF_INFO = 0x00000001

	// Tray icon callback messages (lParam values)
	WM_LBUTTONUP_TRAY   = 0x0202
	WM_LBUTTONDBLCLK    = 0x0203
	WM_RBUTTONUP_TRAY   = 0x0205

	// Menu constants
	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800
	MF_POPUP     = 0x00000010
	MF_CHECKED   = 0x00000008
	MF_GRAYED    = 0x00000001

	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000
	TPM_RETURNCMD   = 0x0100

	// Icon constants
	IMAGE_ICON    = 1
	LR_DEFAULTSIZE = 0x00000040
	LR_LOADFROMFILE = 0x00000010
)

// NOTIFYICONDATAW is the Win32 NOTIFYICONDATAW structure (simplified).
type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

var (
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	procCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	procDestroyMenu      = user32.NewProc("DestroyMenu")
	procAppendMenuW      = user32.NewProc("AppendMenuW")
	procTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procLoadImageW       = user32.NewProc("LoadImageW")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon      = user32.NewProc("DestroyIcon")
	procLoadIconW        = user32.NewProc("LoadIconW")
)

// win32TrayIcon implements platform.TrayIcon using Shell_NotifyIconW.
type win32TrayIcon struct {
	nid           NOTIFYICONDATAW
	menu          *platform.TrayMenu
	menuHandlers  map[uint32]func() // menu item ID -> handler
	onClick       func()
	onDoubleClick func()
	nextMenuID    uint32
	created       bool
	hIcon         uintptr
}

// NewTrayIcon creates a new system tray icon for Windows.
func NewTrayIcon() platform.TrayIcon {
	mw := getMessageWindow()

	t := &win32TrayIcon{
		menuHandlers: make(map[uint32]func()),
		nextMenuID:   1000,
	}

	t.nid.CbSize = uint32(unsafe.Sizeof(NOTIFYICONDATAW{}))
	t.nid.HWnd = mw.Handle()
	t.nid.UID = 1
	t.nid.UFlags = NIF_MESSAGE | NIF_TIP
	t.nid.UCallbackMessage = uint32(WM_APP_TRAY)

	// Set default tooltip
	tip, _ := syscall.UTF16FromString("WindUI")
	copy(t.nid.SzTip[:], tip)

	// Use system default icon (IDI_APPLICATION = 32512)
	hIcon, _, _ := procLoadIconW.Call(0, uintptr(32512))
	if hIcon != 0 {
		t.nid.HIcon = hIcon
		t.nid.UFlags |= NIF_ICON
	}

	// Add the icon
	procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&t.nid)))
	t.created = true

	// Install message handler
	mw.trayHandler = t.handleMessage

	return t
}

func (t *win32TrayIcon) SetIcon(iconData []byte) {
	if len(iconData) == 0 {
		return
	}

	// Create icon from raw ICO data
	hIcon, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&iconData[0])),
		uintptr(len(iconData)),
		1, // fIcon = TRUE
		0x00030000, // version
		0, 0, // default size
		LR_DEFAULTSIZE,
	)

	if hIcon != 0 {
		// Destroy old custom icon
		if t.hIcon != 0 {
			procDestroyIcon.Call(t.hIcon)
		}
		t.hIcon = hIcon
		t.nid.HIcon = hIcon
		t.nid.UFlags |= NIF_ICON
		if t.created {
			procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&t.nid)))
		}
	}
}

func (t *win32TrayIcon) SetTooltip(text string) {
	tip, _ := syscall.UTF16FromString(text)
	copy(t.nid.SzTip[:], tip)
	if t.created {
		t.nid.UFlags |= NIF_TIP
		procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&t.nid)))
	}
}

func (t *win32TrayIcon) SetMenu(menu *platform.TrayMenu) {
	t.menu = menu
}

func (t *win32TrayIcon) ShowBalloon(title, message string) {
	titleUTF16, _ := syscall.UTF16FromString(title)
	msgUTF16, _ := syscall.UTF16FromString(message)

	copy(t.nid.SzInfoTitle[:], titleUTF16)
	copy(t.nid.SzInfo[:], msgUTF16)
	t.nid.UFlags |= NIF_INFO
	t.nid.DwInfoFlags = NIIF_INFO

	if t.created {
		procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&t.nid)))
	}
}

func (t *win32TrayIcon) SetOnClick(fn func()) {
	t.onClick = fn
}

func (t *win32TrayIcon) SetOnDoubleClick(fn func()) {
	t.onDoubleClick = fn
}

func (t *win32TrayIcon) Destroy() {
	if t.created {
		procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&t.nid)))
		t.created = false
	}
	if t.hIcon != 0 {
		procDestroyIcon.Call(t.hIcon)
		t.hIcon = 0
	}
	mw := getMessageWindow()
	mw.trayHandler = nil
}

func (t *win32TrayIcon) handleMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) {
	switch lParam {
	case WM_LBUTTONUP_TRAY:
		if t.onClick != nil {
			t.onClick()
		}

	case WM_LBUTTONDBLCLK:
		if t.onDoubleClick != nil {
			t.onDoubleClick()
		}

	case WM_RBUTTONUP_TRAY:
		t.showContextMenu()
	}
}

func (t *win32TrayIcon) showContextMenu() {
	if t.menu == nil || len(t.menu.Items) == 0 {
		return
	}

	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	// Reset menu handlers
	t.menuHandlers = make(map[uint32]func())
	t.nextMenuID = 1000

	t.buildMenu(hMenu, t.menu.Items)

	// Get cursor position for menu placement
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// Required: SetForegroundWindow before TrackPopupMenu
	mw := getMessageWindow()
	procSetForegroundWindow.Call(mw.Handle())

	// Show the menu and get selected item
	cmd, _, _ := procTrackPopupMenu.Call(
		hMenu,
		TPM_RETURNCMD|TPM_BOTTOMALIGN|TPM_LEFTALIGN,
		uintptr(pt.X), uintptr(pt.Y),
		0,
		mw.Handle(),
		0,
	)

	if cmd != 0 {
		if handler, ok := t.menuHandlers[uint32(cmd)]; ok {
			handler()
		}
	}
}

func (t *win32TrayIcon) buildMenu(hMenu uintptr, items []platform.TrayMenuItem) {
	for _, item := range items {
		if item.IsSeparator {
			procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
			continue
		}

		if len(item.Children) > 0 {
			// Submenu
			hSub, _, _ := procCreatePopupMenu.Call()
			t.buildMenu(hSub, item.Children)

			titlePtr, _ := syscall.UTF16PtrFromString(item.Title)
			procAppendMenuW.Call(hMenu, MF_POPUP, hSub, uintptr(unsafe.Pointer(titlePtr)))
			continue
		}

		// Regular item
		flags := uintptr(MF_STRING)
		if item.Checked {
			flags |= MF_CHECKED
		}
		if item.Disabled {
			flags |= MF_GRAYED
		}

		id := t.nextMenuID
		t.nextMenuID++

		if item.OnClick != nil {
			t.menuHandlers[id] = item.OnClick
		}

		titlePtr, _ := syscall.UTF16PtrFromString(item.Title)
		procAppendMenuW.Call(hMenu, flags, uintptr(id), uintptr(unsafe.Pointer(titlePtr)))
	}
}
