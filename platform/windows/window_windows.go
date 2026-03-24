package windows

import (
	"fmt"
	"image"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"gowui/core"
	"gowui/layout"
	"gowui/platform"
	"gowui/render/gg"
)

// Win32 constants
const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_POPUP            = 0x80000000
	WS_EX_LAYERED       = 0x00080000
	WS_EX_TOPMOST       = 0x00000008

	WM_DESTROY     = 0x0002
	WM_SIZE        = 0x0005
	WM_PAINT       = 0x000F
	WM_CLOSE       = 0x0010
	WM_ERASEBKGND  = 0x0014
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MOUSEWHEEL  = 0x020A
	WM_DPICHANGED  = 0x02E0
	WM_USER        = 0x0400
	WM_APP_PAINT   = WM_USER + 1

	CS_HREDRAW    = 0x0002
	CS_VREDRAW    = 0x0001
	IDC_ARROW     = 32512
	SW_SHOW       = 5
	SW_HIDE       = 0
	CW_USEDEFAULT = ^0x7fffffff
	COLOR_WINDOW  = 5
	SRCCOPY       = 0x00CC0020
	DIB_RGB_COLORS = 0
	BI_RGB         = 0

	SW_MINIMIZE = 6
	SW_MAXIMIZE = 3
	SW_RESTORE  = 9

	SWP_NOMOVE   = 0x0002
	SWP_NOSIZE   = 0x0001
	SWP_NOZORDER = 0x0004

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	GWL_STYLE = -16

	WS_MAXIMIZE = 0x01000000

	HWND_TOPMOST   = ^uintptr(0) // -1 as uintptr
	HWND_NOTOPMOST = ^uintptr(1) // -2 as uintptr
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procPostMessageW      = user32.NewProc("PostMessageW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procGetClientRect     = user32.NewProc("GetClientRect")
	procInvalidateRect    = user32.NewProc("InvalidateRect")
	procBeginPaint        = user32.NewProc("BeginPaint")
	procEndPaint          = user32.NewProc("EndPaint")
	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procGetSystemMetrics  = user32.NewProc("GetSystemMetrics")
	procMoveWindow        = user32.NewProc("MoveWindow")
	procLoadCursorW       = user32.NewProc("LoadCursorW")
	procSetWindowTextW    = user32.NewProc("SetWindowTextW")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procIsWindowVisible   = user32.NewProc("IsWindowVisible")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procGetWindowLongW    = user32.NewProc("GetWindowLongW")
	procScreenToClient    = user32.NewProc("ScreenToClient")
	procIsDialogMessageW  = user32.NewProc("IsDialogMessageW")

	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procBitBlt             = gdi32.NewProc("BitBlt")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
	procDeleteObject       = gdi32.NewProc("DeleteObject")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

// WNDCLASSEXW is the Windows WNDCLASSEXW structure.
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// MSG is the Windows MSG structure.
type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// POINT is the Windows POINT structure.
type POINT struct {
	X, Y int32
}

// RECT is the Windows RECT structure.
type RECT struct {
	Left, Top, Right, Bottom int32
}

// PAINTSTRUCT is the Windows PAINTSTRUCT structure.
type PAINTSTRUCT struct {
	HDC         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

// BITMAPINFOHEADER is the Windows BITMAPINFOHEADER structure.
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// BITMAPINFO is the Windows BITMAPINFO structure.
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

// windowMap stores HWND -> *win32Window mapping for wndProc lookups.
var windowMap sync.Map

// classRegistered tracks whether the window class has been registered.
var classRegistered bool
var className = syscall.StringToUTF16Ptr("GoWUIWindowClass")

// win32Window implements platform.Window for Windows.
type win32Window struct {
	hwnd     uintptr
	plat     *WindowsPlatform
	opts     platform.WindowOptions

	contentView  *core.Node
	textRenderer core.TextRenderer
	dpiScale     float64 // DPI scale factor (1.0 at 96 DPI, 1.5 at 144 DPI)
	dpiScaled    bool    // true after node tree has been DPI-scaled

	lastHoverNode   *core.Node // tracks which node the mouse is over for HoverEnter/Exit
	capturedNode    *core.Node // pointer capture: receives all Move/Up events after ActionDown

	onClose        func() bool
	onResize       func(w, h int)
	onDPIChanged   func(dpi float64)
	onFocusChanged func(focused bool)

	// Cached rendering buffers — reused across frames, recreated on resize only.
	cachedImage *image.RGBA     // canvas backing buffer (avoids re-alloc per frame)
	dibMemDC    uintptr         // cached memory DC for presentation
	dibBitmap   uintptr         // cached DIB section HBITMAP
	dibBits     unsafe.Pointer  // pointer to DIB pixel data
	dibWidth    int             // cached DIB width
	dibHeight   int             // cached DIB height

	mu sync.Mutex
}

// registerWindowClass registers the Win32 window class once.
func registerWindowClass() error {
	if classRegistered {
		return nil
	}

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	hCursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)

	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       hCursor,
		HbrBackground: COLOR_WINDOW + 1,
		LpszClassName: className,
	}

	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return err
	}
	classRegistered = true
	return nil
}

// newWin32Window creates a new Win32 window.
func newWin32Window(plat *WindowsPlatform, opts platform.WindowOptions) (*win32Window, error) {
	runtime.LockOSThread()

	if err := registerWindowClass(); err != nil {
		return nil, err
	}

	w := &win32Window{
		plat: plat,
		opts: opts,
	}

	hInstance, _, _ := procGetModuleHandleW.Call(0)

	// Determine window styles
	// WS_CLIPCHILDREN excludes child window areas from parent painting,
	// so native EDIT controls aren't covered by our BitBlt.
	style := uintptr(WS_OVERLAPPEDWINDOW | 0x02000000) // WS_CLIPCHILDREN = 0x02000000
	exStyle := uintptr(0)

	if opts.Frameless {
		style = WS_POPUP | WS_VISIBLE
	}
	if opts.TopMost {
		exStyle |= WS_EX_TOPMOST
	}
	if opts.Transparent {
		exStyle |= WS_EX_LAYERED
	}

	// Determine position
	x := CW_USEDEFAULT
	y := CW_USEDEFAULT
	if opts.X != 0 || opts.Y != 0 {
		x = opts.X
		y = opts.Y
	}

	// Determine size — scale dp to physical pixels using system DPI.
	// After DPI awareness is enabled, CreateWindowExW expects physical pixels.
	sysDPI := 96.0
	if procGetDpiForWindow.Find() == nil {
		// Before window exists, use GetDpiForSystem or default
		getDpiForSystem := syscall.NewLazyDLL("user32.dll").NewProc("GetDpiForSystem")
		if getDpiForSystem.Find() == nil {
			d, _, _ := getDpiForSystem.Call()
			if d > 0 {
				sysDPI = float64(d)
			}
		}
	}
	width := opts.Width
	height := opts.Height
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 600
	}
	// Scale from dp to physical pixels
	width = int(DpToPx(float64(width), sysDPI))
	height = int(DpToPx(float64(height), sysDPI))

	title := syscall.StringToUTF16Ptr(opts.Title)

	hwnd, _, err := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		style,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		0, // parent
		0, // menu
		hInstance,
		0, // lpParam
	)
	if hwnd == 0 {
		return nil, err
	}

	w.hwnd = hwnd
	windowMap.Store(hwnd, w)

	return w, nil
}

// wndProc is the Win32 window procedure callback.
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	val, ok := windowMap.Load(hwnd)
	if !ok {
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
	w := val.(*win32Window)

	switch msg {
	case WM_CLOSE:
		if w.onClose != nil {
			if !w.onClose() {
				return 0 // Prevent close
			}
		}
		procDestroyWindow.Call(hwnd)
		return 0

	case WM_DESTROY:
		w.releaseCachedDIB()
		w.cachedImage = nil
		windowMap.Delete(hwnd)
		// Post WM_QUIT to exit the message loop
		procPostQuitMessage.Call(0)
		return 0

	case WM_DPICHANGED:
		// wParam: LOWORD = new X DPI, HIWORD = new Y DPI
		newDPI := float64(int16(wParam & 0xFFFF))
		newScale := newDPI / 96.0
		if newScale > 0 && newScale != w.dpiScale {
			oldScale := w.dpiScale
			w.dpiScale = newScale
			// Re-scale the node tree by the ratio of new/old DPI
			if w.contentView != nil && oldScale > 0 {
				ratio := newScale / oldScale
				rescaleNodeTree(w.contentView, ratio)
			}
			// lParam points to a RECT with the suggested new window position/size
			if lParam != 0 {
				suggestedRect := (*RECT)(unsafe.Pointer(lParam))
				procSetWindowPos.Call(hwnd, 0,
					uintptr(suggestedRect.Left), uintptr(suggestedRect.Top),
					uintptr(suggestedRect.Right-suggestedRect.Left),
					uintptr(suggestedRect.Bottom-suggestedRect.Top),
					SWP_NOZORDER)
			}
			if w.onDPIChanged != nil {
				w.onDPIChanged(newDPI)
			}
		}
		return 0

	case WM_SIZE:
		width := int(lParam & 0xFFFF)
		height := int((lParam >> 16) & 0xFFFF)
		if w.onResize != nil {
			w.onResize(width, height)
		}
		w.render()
		return 0

	case WM_PAINT:
		var ps PAINTSTRUCT
		procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		w.render()
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_APP_PAINT:
		w.render()
		return 0

	case WM_ERASEBKGND:
		return 1 // We handle background painting ourselves

	case WM_MOUSEMOVE:
		x := int(int16(lParam & 0xFFFF))
		y := int(int16((lParam >> 16) & 0xFFFF))
		// MK_LBUTTON (0x0001) indicates left button is held during move
		if wParam&0x0001 != 0 {
			w.dispatchMotion(core.ActionMove, float64(x), float64(y), core.MouseButtonLeft)
		} else {
			w.dispatchMotion(core.ActionHoverMove, float64(x), float64(y), core.MouseButtonLeft)
		}
		return 0

	case WM_LBUTTONDOWN:
		x := int(int16(lParam & 0xFFFF))
		y := int(int16((lParam >> 16) & 0xFFFF))
		w.dispatchMotion(core.ActionDown, float64(x), float64(y), core.MouseButtonLeft)
		return 0

	case WM_LBUTTONUP:
		x := int(int16(lParam & 0xFFFF))
		y := int(int16((lParam >> 16) & 0xFFFF))
		w.dispatchMotion(core.ActionUp, float64(x), float64(y), core.MouseButtonLeft)
		return 0

	case WM_RBUTTONDOWN:
		x := int(int16(lParam & 0xFFFF))
		y := int(int16((lParam >> 16) & 0xFFFF))
		w.dispatchMotion(core.ActionDown, float64(x), float64(y), core.MouseButtonRight)
		return 0

	case WM_RBUTTONUP:
		x := int(int16(lParam & 0xFFFF))
		y := int(int16((lParam >> 16) & 0xFFFF))
		w.dispatchMotion(core.ActionUp, float64(x), float64(y), core.MouseButtonRight)
		return 0

	case WM_MOUSEWHEEL:
		// wParam high word is wheel delta (positive = scroll up, negative = scroll down)
		delta := int16(wParam >> 16)
		// WHEEL_DELTA is 120; normalize to notch count
		notches := float64(delta) / 120.0
		// lParam contains SCREEN coordinates — convert to client coordinates
		pt := POINT{X: int32(int16(lParam & 0xFFFF)), Y: int32(int16((lParam >> 16) & 0xFFFF))}
		procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
		w.dispatchScroll(float64(pt.X), float64(pt.Y), 0, notches)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// dispatchMotion creates and dispatches a MotionEvent through the node tree
// using the 3-phase dispatch (hit-test → intercept → handle+bubble).
func (w *win32Window) dispatchMotion(action core.MotionAction, x, y float64, button core.MouseButton) {
	if w.contentView == nil {
		return
	}

	// Hover tracking: send HoverEnter/HoverExit when the deepest hit node changes.
	// HoverExit bubbles UP to all ancestors so parent containers (ScrollView etc.)
	// can clear their hover state when the mouse leaves their bounds.
	if action == core.ActionHoverMove {
		hitPoint := core.Point{X: x, Y: y}
		currentTarget := core.HitTest(w.contentView, hitPoint)
		if currentTarget != w.lastHoverNode {
			needRepaint := false
			// Send HoverExit to old target AND its ancestors
			if w.lastHoverNode != nil {
				exitEvt := core.NewMotionEvent(core.ActionHoverExit, x, y)
				for node := w.lastHoverNode; node != nil; node = node.Parent() {
					if h := node.GetHandler(); h != nil {
						if h.OnEvent(node, exitEvt) {
							needRepaint = true
						}
					}
				}
			}
			// Send HoverEnter to new target only (not ancestors)
			if currentTarget != nil {
				enterEvt := core.NewMotionEvent(core.ActionHoverEnter, x, y)
				if h := currentTarget.GetHandler(); h != nil {
					if h.OnEvent(currentTarget, enterEvt) {
						needRepaint = true
					}
				}
			}
			w.lastHoverNode = currentTarget
			if needRepaint {
				w.Invalidate()
			}
		}
	}

	// Pointer capture: after ActionDown, subsequent Move/Up go to the node
	// that consumed ActionDown, so drag works even outside the original view.
	if (action == core.ActionMove || action == core.ActionUp) && w.capturedNode != nil {
		evt := core.NewMotionEvent(action, x, y)
		evt.Button = button
		evt.RawX = x
		evt.RawY = y
		consumed := false
		if h := w.capturedNode.GetHandler(); h != nil {
			consumed = h.OnEvent(w.capturedNode, evt)
		}
		if action == core.ActionUp {
			w.capturedNode = nil
		}
		if consumed {
			w.Invalidate()
		}
		return
	}

	evt := core.NewMotionEvent(action, x, y)
	evt.Button = button
	evt.RawX = x
	evt.RawY = y

	// Dispatch through the normal 3-phase event system.
	// For ActionDown, capture the consuming node for subsequent Move/Up.
	if action == core.ActionDown {
		consumer, consumed := core.DispatchEventCapture(w.contentView, evt, core.Point{X: x, Y: y})
		w.capturedNode = consumer
		if consumed {
			w.Invalidate()
		}
	} else {
		consumed := core.DispatchEvent(w.contentView, evt, core.Point{X: x, Y: y})
		if consumed {
			w.Invalidate()
		}
	}
}

// dispatchScroll creates and dispatches a ScrollEvent through the node tree.
func (w *win32Window) dispatchScroll(x, y, deltaX, deltaY float64) {
	if w.contentView == nil {
		return
	}
	evt := core.NewScrollEvent(x, y, deltaX, deltaY)
	consumed := core.DispatchEvent(w.contentView, evt, core.Point{X: x, Y: y})
	if consumed {
		w.Invalidate() // repaint after scroll
	}
}

// ---------- platform.Window implementation ----------

func (w *win32Window) SetContentView(root *core.Node) {
	w.mu.Lock()
	w.contentView = root
	w.mu.Unlock()
	w.Invalidate()
}

func (w *win32Window) SetTitle(title string) {
	t := syscall.StringToUTF16Ptr(title)
	procSetWindowTextW.Call(w.hwnd, uintptr(unsafe.Pointer(t)))
}

func (w *win32Window) SetIcon(icon *core.ImageResource) {
	// Phase 1: stub — setting window icon requires CreateIconFromResourceEx
}

func (w *win32Window) Show() {
	procShowWindow.Call(w.hwnd, SW_SHOW)
	// Force immediate render after showing (ShowWindow triggers WM_PAINT,
	// but we also render explicitly to ensure content is visible immediately)
	w.render()
}

func (w *win32Window) Hide() {
	procShowWindow.Call(w.hwnd, SW_HIDE)
}

func (w *win32Window) Close() {
	procDestroyWindow.Call(w.hwnd)
}

func (w *win32Window) Minimize() {
	procShowWindow.Call(w.hwnd, SW_MINIMIZE)
}

func (w *win32Window) Maximize() {
	procShowWindow.Call(w.hwnd, SW_MAXIMIZE)
}

func (w *win32Window) Restore() {
	procShowWindow.Call(w.hwnd, SW_RESTORE)
}

func (w *win32Window) SetSize(width, height int) {
	procSetWindowPos.Call(w.hwnd, 0,
		0, 0, uintptr(width), uintptr(height),
		SWP_NOMOVE|SWP_NOZORDER)
}

func (w *win32Window) SetPosition(x, y int) {
	procSetWindowPos.Call(w.hwnd, 0,
		uintptr(x), uintptr(y), 0, 0,
		SWP_NOSIZE|SWP_NOZORDER)
}

func (w *win32Window) Center() {
	width, height := w.getWindowSize()
	screenW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	x := (int(screenW) - width) / 2
	y := (int(screenH) - height) / 2
	w.SetPosition(x, y)
}

func (w *win32Window) IsVisible() bool {
	ret, _, _ := procIsWindowVisible.Call(w.hwnd)
	return ret != 0
}

func (w *win32Window) IsFocused() bool {
	fg, _, _ := procGetForegroundWindow.Call()
	return fg == w.hwnd
}

func (w *win32Window) GetSize() core.Size {
	width, height := w.getClientSize()
	return core.Size{Width: float64(width), Height: float64(height)}
}

func (w *win32Window) GetPosition() core.Point {
	var r RECT
	procGetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&r)))
	return core.Point{X: float64(r.Left), Y: float64(r.Top)}
}

var procGetDpiForWindow *syscall.LazyProc

func init() {
	procGetDpiForWindow = syscall.NewLazyDLL("user32.dll").NewProc("GetDpiForWindow")
}

func (w *win32Window) GetDPI() float64 {
	if procGetDpiForWindow.Find() == nil {
		dpi, _, _ := procGetDpiForWindow.Call(w.hwnd)
		if dpi > 0 {
			return float64(dpi)
		}
	}
	return 96.0
}

func (w *win32Window) SetOnClose(fn func() bool) {
	w.onClose = fn
}

func (w *win32Window) SetOnResize(fn func(w, h int)) {
	w.onResize = fn
}

func (w *win32Window) SetOnDPIChanged(fn func(dpi float64)) {
	w.onDPIChanged = fn
}

func (w *win32Window) SetOnFocusChanged(fn func(focused bool)) {
	w.onFocusChanged = fn
}

func (w *win32Window) NativeHandle() uintptr {
	return w.hwnd
}

func (w *win32Window) Invalidate() {
	procPostMessageW.Call(w.hwnd, WM_APP_PAINT, 0, 0)
}

func (w *win32Window) InvalidateRect(rect core.Rect) {
	// Phase 1: invalidate entire window
	w.Invalidate()
}

// ---------- Render pipeline ----------

func (w *win32Window) render() {
	w.mu.Lock()
	contentView := w.contentView
	w.mu.Unlock()

	if contentView == nil {
		return
	}

	width, height := w.getClientSize()
	if width <= 0 || height <= 0 {
		return
	}

	// Update DPI scale factor
	w.dpiScale = w.GetDPI() / 96.0
	if w.dpiScale < 1.0 {
		w.dpiScale = 1.0
	}

	// Scale dp values in the node tree to physical pixels (once).
	// Uses core.ScaleNodeDPI which marks nodes as scaled, so dynamically
	// added children via AddChild will also be auto-scaled.
	if !w.dpiScaled {
		core.ScaleNodeDPI(contentView, w.dpiScale)
		w.dpiScaled = true
	}

	// Create canvas with cached text renderer
	if w.textRenderer == nil {
		w.textRenderer = w.plat.CreateTextRenderer()
	}

	// Reuse the canvas backing image if the window size hasn't changed.
	// Only the lightweight gg.Context is recreated; the large pixel buffer
	// (~width*height*4 bytes) is kept across frames.
	var canvas *gg.GGCanvas
	if w.cachedImage != nil &&
		w.cachedImage.Bounds().Dx() == width &&
		w.cachedImage.Bounds().Dy() == height {
		canvas = gg.NewGGCanvasForImage(w.cachedImage, w.textRenderer)
	} else {
		fmt.Printf("[GoWUI] canvas resized: %dx%d (%.1f MB RGBA)\n",
			width, height, float64(width*height*4)/(1024*1024))
		canvas = gg.NewGGCanvas(width, height, w.textRenderer)
		w.cachedImage = canvas.Target()
	}

	// Measure
	root := contentView
	widthSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(width)}
	heightSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(height)}
	layout.MeasureChild(root, widthSpec, heightSpec)

	// Arrange
	if l := root.GetLayout(); l != nil {
		l.Arrange(root, core.Rect{Width: float64(width), Height: float64(height)})
	}
	root.SetBounds(core.Rect{Width: float64(width), Height: float64(height)})

	// Paint
	PaintNode(root, canvas)

	// Present: BitBlt canvas to window
	w.present(canvas.Target())
}

// rescaleNodeTree re-scales all nodes by a ratio (newDPI/oldDPI) when the
// display DPI changes at runtime. It resets the dpiScaled flag and applies
// the ratio to all dp-valued fields.
func rescaleNodeTree(node *core.Node, ratio float64) {
	// Update stored dpiScale
	if s, ok := node.GetData("dpiScale").(float64); ok {
		node.SetData("dpiScale", s*ratio)
	}

	// Scale padding
	p := node.Padding()
	node.SetPadding(core.Insets{
		Left: p.Left * ratio, Top: p.Top * ratio,
		Right: p.Right * ratio, Bottom: p.Bottom * ratio,
	})

	// Scale margin
	m := node.Margin()
	node.SetMargin(core.Insets{
		Left: m.Left * ratio, Top: m.Top * ratio,
		Right: m.Right * ratio, Bottom: m.Bottom * ratio,
	})

	// Scale style (already in physical px from previous scaling, multiply by ratio)
	if s := node.GetStyle(); s != nil {
		if s.FontSize > 0 {
			s.FontSize *= ratio
		}
		if s.CornerRadius > 0 {
			s.CornerRadius *= ratio
		}
		if s.BorderWidth > 0 {
			s.BorderWidth *= ratio
		}
		if s.Width.Unit == core.DimensionDp && s.Width.Value > 0 {
			s.Width.Value *= ratio
		}
		if s.Height.Unit == core.DimensionDp && s.Height.Value > 0 {
			s.Height.Value *= ratio
		}
	}

	// Scale layout spacing
	if l := node.GetLayout(); l != nil {
		if ds, ok := l.(core.DPIScalable); ok {
			ds.ScaleDPI(ratio)
		}
	}

	for _, child := range node.Children() {
		rescaleNodeTree(child, ratio)
	}
}

// PaintNode recursively paints a node tree onto a canvas.
// If a node has the "paintsChildren" data flag set, child painting is
// skipped here because the node's Painter handles it internally (e.g.
// ScrollView applies clipping and scroll-offset translation).
func PaintNode(node *core.Node, canvas core.Canvas) {
	if node.GetVisibility() != core.Visible {
		return
	}
	canvas.Save()
	b := node.Bounds()
	canvas.Translate(b.X, b.Y)
	if p := node.GetPainter(); p != nil {
		p.Paint(node, canvas)
	}
	// Skip child painting if the painter already handled it
	if node.GetData("paintsChildren") == nil {
		for _, child := range node.Children() {
			PaintNode(child, canvas)
		}
	}
	canvas.Restore()
}

// present copies the RGBA image to the window using GDI.
// The memory DC and DIB section are cached and reused across frames;
// they are only recreated when the window size changes.
func (w *win32Window) present(img *image.RGBA) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	hdc, _, _ := procGetDC.Call(w.hwnd)
	if hdc == 0 {
		return
	}
	defer procReleaseDC.Call(w.hwnd, hdc)

	// Ensure cached DIB section matches the current dimensions.
	if w.dibWidth != width || w.dibHeight != height {
		fmt.Printf("[GoWUI] DIB resized: %dx%d (%.1f MB BGRA)\n",
			width, height, float64(width*height*4)/(1024*1024))
		w.releaseCachedDIB()

		memDC, _, _ := procCreateCompatibleDC.Call(hdc)
		if memDC == 0 {
			return
		}

		bmi := BITMAPINFO{
			BmiHeader: BITMAPINFOHEADER{
				BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
				BiWidth:       int32(width),
				BiHeight:      -int32(height), // top-down
				BiPlanes:      1,
				BiBitCount:    32,
				BiCompression: BI_RGB,
			},
		}

		var bits unsafe.Pointer
		hBitmap, _, _ := procCreateDIBSection.Call(
			memDC,
			uintptr(unsafe.Pointer(&bmi)),
			DIB_RGB_COLORS,
			uintptr(unsafe.Pointer(&bits)),
			0,
			0,
		)
		if hBitmap == 0 {
			procDeleteDC.Call(memDC)
			return
		}

		procSelectObject.Call(memDC, hBitmap)

		w.dibMemDC = memDC
		w.dibBitmap = hBitmap
		w.dibBits = bits
		w.dibWidth = width
		w.dibHeight = height
	}

	// Copy pixels with RGBA -> BGRA conversion
	pix := img.Pix
	stride := img.Stride
	dibStride := width * 4

	dst := unsafe.Slice((*byte)(w.dibBits), height*dibStride)
	for y := 0; y < height; y++ {
		srcOff := y * stride
		dstOff := y * dibStride
		for x := 0; x < width; x++ {
			si := srcOff + x*4
			di := dstOff + x*4
			// RGBA -> BGRA: swap R and B
			dst[di+0] = pix[si+2] // B
			dst[di+1] = pix[si+1] // G
			dst[di+2] = pix[si+0] // R
			dst[di+3] = pix[si+3] // A
		}
	}

	// BitBlt to window
	procBitBlt.Call(
		hdc,
		0, 0,
		uintptr(width), uintptr(height),
		w.dibMemDC,
		0, 0,
		SRCCOPY,
	)
}

// releaseCachedDIB frees the cached GDI resources (memory DC + DIB section).
func (w *win32Window) releaseCachedDIB() {
	if w.dibBitmap != 0 {
		procDeleteObject.Call(w.dibBitmap)
		w.dibBitmap = 0
	}
	if w.dibMemDC != 0 {
		procDeleteDC.Call(w.dibMemDC)
		w.dibMemDC = 0
	}
	w.dibBits = nil
	w.dibWidth = 0
	w.dibHeight = 0
}

// ---------- Helpers ----------

// getClientSize returns the client area dimensions.
func (w *win32Window) getClientSize() (int, int) {
	var r RECT
	procGetClientRect.Call(w.hwnd, uintptr(unsafe.Pointer(&r)))
	return int(r.Right - r.Left), int(r.Bottom - r.Top)
}

// getWindowSize returns the full window dimensions (including frame).
func (w *win32Window) getWindowSize() (int, int) {
	var r RECT
	procGetWindowRect.Call(w.hwnd, uintptr(unsafe.Pointer(&r)))
	return int(r.Right - r.Left), int(r.Bottom - r.Top)
}
