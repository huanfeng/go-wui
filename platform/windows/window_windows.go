package windows

import (
	"image"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"gowui/core"
	"gowui/layout"
	"gowui/platform"
	"gowui/render/freetype"
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

	contentView *core.Node

	onClose        func() bool
	onResize       func(w, h int)
	onDPIChanged   func(dpi float64)
	onFocusChanged func(focused bool)

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
	style := uintptr(WS_OVERLAPPEDWINDOW)
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

	// Determine size
	width := opts.Width
	height := opts.Height
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 600
	}

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
		windowMap.Delete(hwnd)
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
		w.dispatchMotion(core.ActionMove, float64(x), float64(y), core.MouseButtonLeft)
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
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// dispatchMotion creates and dispatches a MotionEvent to the content view.
func (w *win32Window) dispatchMotion(action core.MotionAction, x, y float64, button core.MouseButton) {
	if w.contentView == nil {
		return
	}
	evt := core.NewMotionEvent(action, x, y)
	evt.Button = button
	evt.RawX = x
	evt.RawY = y

	handler := w.contentView.GetHandler()
	if handler != nil {
		handler.OnDispatchEvent(w.contentView, evt)
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

func (w *win32Window) GetDPI() float64 {
	// Phase 1: return system default DPI
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

	// Create canvas with text renderer
	textRenderer := freetype.NewFreeTypeTextRenderer()
	canvas := gg.NewGGCanvas(width, height, textRenderer)

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

// PaintNode recursively paints a node tree onto a canvas.
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
	for _, child := range node.Children() {
		PaintNode(child, canvas)
	}
	canvas.Restore()
}

// present copies the RGBA image to the window using GDI.
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

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return
	}
	defer procDeleteDC.Call(memDC)

	// Create DIB section
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
		return
	}
	defer procDeleteObject.Call(hBitmap)

	procSelectObject.Call(memDC, hBitmap)

	// Copy pixels with RGBA -> BGRA conversion
	pix := img.Pix
	stride := img.Stride
	dibStride := width * 4

	dst := unsafe.Slice((*byte)(bits), height*dibStride)
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
		memDC,
		0, 0,
		SRCCOPY,
	)
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
