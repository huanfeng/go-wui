//go:build windows

package windows

import (
	"image"
	"math"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/huanfeng/wind-ui/core"
)

// GUID for COM
type comGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// DirectWrite constants
const (
	DWRITE_FACTORY_TYPE_SHARED      = 0
	DWRITE_FONT_WEIGHT_NORMAL      = 400
	DWRITE_FONT_STYLE_NORMAL       = 0
	DWRITE_FONT_STRETCH_NORMAL     = 5
	DWRITE_WORD_WRAPPING_WRAP      = 0
	DWRITE_WORD_WRAPPING_NO_WRAP   = 1
	DWRITE_TEXT_ALIGNMENT_LEADING  = 0
	DWRITE_PARAGRAPH_ALIGNMENT_NEAR = 0
)

var iidIDWriteFactory = comGUID{
	0xb859ee5a, 0xd838, 0x4b5b,
	[8]byte{0xa2, 0xe8, 0x1a, 0xdc, 0x7d, 0x93, 0xdb, 0x48},
}

// GDI constants for bitmap readback.
const (
	_OBJ_BITMAP    = 7
	_BLACKNESS     = 0x00000042
	_DWRITE_MEASURING_MODE_NATURAL = 0
	_S_OK          = 0
)

// DLL and proc
var (
	dwriteDll               *syscall.LazyDLL
	procDWriteCreateFactory *syscall.LazyProc
	dwriteInitOnce          sync.Once
	dwriteInitErr           error

	// Additional GDI procs for bitmap readback.
	procGetDIBits        *syscall.LazyProc
	procSetDIBits        *syscall.LazyProc
	procGetCurrentObject *syscall.LazyProc
)

func initDWrite() error {
	dwriteInitOnce.Do(func() {
		dwriteDll = syscall.NewLazyDLL("dwrite.dll")
		procDWriteCreateFactory = dwriteDll.NewProc("DWriteCreateFactory")
		dwriteInitErr = procDWriteCreateFactory.Find()
		if dwriteInitErr != nil {
			return
		}
		// Initialize GDI procs needed for bitmap readback.
		gdi := syscall.NewLazyDLL("gdi32.dll")
		procGetDIBits = gdi.NewProc("GetDIBits")
		procSetDIBits = gdi.NewProc("SetDIBits")
		procGetCurrentObject = gdi.NewProc("GetCurrentObject")
	})
	return dwriteInitErr
}

// comCall calls a COM method via vtable index.
func comCall(obj uintptr, vtableIndex int, args ...uintptr) (uintptr, error) {
	// obj is pointer to pointer to vtable
	vtablePtr := *(*uintptr)(unsafe.Pointer(obj))
	methodPtr := *(*uintptr)(unsafe.Pointer(vtablePtr + uintptr(vtableIndex)*unsafe.Sizeof(uintptr(0))))

	allArgs := make([]uintptr, 0, 1+len(args))
	allArgs = append(allArgs, obj) // "this" pointer
	allArgs = append(allArgs, args...)

	ret, _, _ := syscall.SyscallN(methodPtr, allArgs...)
	if int32(ret) < 0 { // HRESULT is negative = error
		return ret, syscall.Errno(ret)
	}
	return ret, nil
}

// comRelease calls IUnknown::Release (vtable index 2).
func comRelease(obj uintptr) {
	if obj != 0 {
		comCall(obj, 2) // Release is always at index 2
	}
}

// DWRITE_TEXT_METRICS structure.
type dwriteTextMetrics struct {
	Left                             float32
	Top                              float32
	Width                            float32
	WidthIncludingTrailingWhitespace float32
	Height                           float32
	LayoutWidth                      float32
	LayoutHeight                     float32
	MaxBidiReorderingDepth           uint32
	LineCount                        uint32
}

// DWRITE_LINE_METRICS structure.
type dwriteLineMetrics struct {
	Length                   uint32
	TrailingWhitespaceLength uint32
	NewlineLength            uint32
	Height                   float32
	Baseline                 float32
	IsTrimmed                int32 // BOOL
}

// ---------- GDI Interop backend ----------

// gdiInteropBackend implements textDrawBackend using IDWriteGdiInterop +
// IDWriteBitmapRenderTarget — no D2D/D3D dependency.
type gdiInteropBackend struct {
	gdiInterop   uintptr // IDWriteGdiInterop*
	bitmapTarget uintptr // IDWriteBitmapRenderTarget*
	renderParams uintptr // IDWriteRenderingParams*
	width, height int

	// Custom text renderer COM object for IDWriteTextLayout::Draw callback.
	renderer *goTextRenderer
}

// goTextRendererVtable is the COM vtable for our IDWriteTextRenderer implementation.
// Layout: IUnknown (3) + IDWritePixelSnapping (3) + IDWriteTextRenderer (4) = 10 methods.
type goTextRendererVtable struct {
	QueryInterface          uintptr
	AddRef                  uintptr
	Release                 uintptr
	IsPixelSnappingDisabled uintptr
	GetCurrentTransform     uintptr
	GetPixelsPerDip         uintptr
	DrawGlyphRun            uintptr
	DrawUnderline           uintptr
	DrawStrikethrough       uintptr
	DrawInlineObject        uintptr
}

// goTextRenderer is our Go-implemented IDWriteTextRenderer COM object.
// The vtable pointer MUST be the first field (COM convention: object pointer IS
// pointer-to-vtable-pointer).
type goTextRenderer struct {
	vtable       *goTextRendererVtable // must be first field
	refCount     int32
	bitmapTarget uintptr // IDWriteBitmapRenderTarget* — delegate DrawGlyphRun here
	renderParams uintptr // IDWriteRenderingParams*
	textColor    uint32  // COLORREF (0x00BBGGRR)
	drawCallCount int32   // diagnostic: counts DrawGlyphRun invocations
	lastDrawHR    uintptr // diagnostic: last HRESULT from DrawGlyphRun
}

// DWRITE_MATRIX is used by GetCurrentTransform (identity matrix).
type dwriteMatrix struct {
	M11 float32
	M12 float32
	M21 float32
	M22 float32
	Dx  float32
	Dy  float32
}

// Global vtable instance — initialized once, shared by all goTextRenderer instances.
var (
	globalTextRendererVtable     *goTextRendererVtable
	globalTextRendererVtableOnce sync.Once
)

func initGoTextRendererVtable() *goTextRendererVtable {
	globalTextRendererVtableOnce.Do(func() {
		globalTextRendererVtable = &goTextRendererVtable{
			QueryInterface:          syscall.NewCallback(goTR_QueryInterface),
			AddRef:                  syscall.NewCallback(goTR_AddRef),
			Release:                 syscall.NewCallback(goTR_Release),
			IsPixelSnappingDisabled: syscall.NewCallback(goTR_IsPixelSnappingDisabled),
			GetCurrentTransform:     syscall.NewCallback(goTR_GetCurrentTransform),
			GetPixelsPerDip:         syscall.NewCallback(goTR_GetPixelsPerDip),
			DrawGlyphRun:            cgoDrawGlyphRunCallback(), // CGO bridge — correct float ABI via XMM registers
			DrawUnderline:           syscall.NewCallback(goTR_DrawUnderline),
			DrawStrikethrough:       syscall.NewCallback(goTR_DrawStrikethrough),
			DrawInlineObject:        syscall.NewCallback(goTR_DrawInlineObject),
		}
	})
	return globalTextRendererVtable
}

// --- COM callback implementations ---

// IID constants for QueryInterface.
var (
	_IID_IUnknown = comGUID{
		0x00000000, 0x0000, 0x0000,
		[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	_IID_IDWriteTextRenderer = comGUID{
		0xef8a8135, 0x5cc6, 0x45fe,
		[8]byte{0x88, 0x25, 0xc5, 0xa0, 0x72, 0x4e, 0xb8, 0x19},
	}
	_IID_IDWritePixelSnapping = comGUID{
		0xeaf3a2da, 0xecf4, 0x4d24,
		[8]byte{0xb6, 0x44, 0xb3, 0x4f, 0x68, 0x42, 0x02, 0x4b},
	}
)

func guidEqual(a, b *comGUID) bool {
	return a.Data1 == b.Data1 && a.Data2 == b.Data2 && a.Data3 == b.Data3 && a.Data4 == b.Data4
}

// goTR_QueryInterface — IUnknown::QueryInterface.
// We support IUnknown, IDWritePixelSnapping, and IDWriteTextRenderer.
func goTR_QueryInterface(this uintptr, riid uintptr, ppvObject uintptr) uintptr {
	if ppvObject == 0 {
		return 0x80004003 // E_POINTER
	}

	iid := (*comGUID)(unsafe.Pointer(riid))
	if guidEqual(iid, &_IID_IUnknown) ||
		guidEqual(iid, &_IID_IDWriteTextRenderer) ||
		guidEqual(iid, &_IID_IDWritePixelSnapping) {
		*(*uintptr)(unsafe.Pointer(ppvObject)) = this
		goTR_AddRef(this)
		return _S_OK
	}

	*(*uintptr)(unsafe.Pointer(ppvObject)) = 0
	return 0x80004002 // E_NOINTERFACE
}

// goTR_AddRef — IUnknown::AddRef.
func goTR_AddRef(this uintptr) uintptr {
	tr := (*goTextRenderer)(unsafe.Pointer(this))
	return uintptr(atomic.AddInt32(&tr.refCount, 1))
}

// goTR_Release — IUnknown::Release.
func goTR_Release(this uintptr) uintptr {
	tr := (*goTextRenderer)(unsafe.Pointer(this))
	n := atomic.AddInt32(&tr.refCount, -1)
	// We manage the lifetime externally, so don't free here.
	return uintptr(n)
}

// goTR_IsPixelSnappingDisabled — IDWritePixelSnapping method.
func goTR_IsPixelSnappingDisabled(this uintptr, clientDrawingContext uintptr, isDisabled uintptr) uintptr {
	if isDisabled != 0 {
		*(*int32)(unsafe.Pointer(isDisabled)) = 0 // FALSE — pixel snapping enabled
	}
	return _S_OK
}

// goTR_GetCurrentTransform — IDWritePixelSnapping method.
func goTR_GetCurrentTransform(this uintptr, clientDrawingContext uintptr, transform uintptr) uintptr {
	if transform != 0 {
		m := (*dwriteMatrix)(unsafe.Pointer(transform))
		*m = dwriteMatrix{M11: 1.0, M22: 1.0} // identity
	}
	return _S_OK
}

// goTR_GetPixelsPerDip — IDWritePixelSnapping method.
func goTR_GetPixelsPerDip(this uintptr, clientDrawingContext uintptr, pixelsPerDip uintptr) uintptr {
	if pixelsPerDip != 0 {
		*(*float32)(unsafe.Pointer(pixelsPerDip)) = 1.0
	}
	return _S_OK
}

// goTR_DrawGlyphRun is now handled by the CGO trampoline in dwrite_cgo_windows.go.
// The C bridge correctly receives float parameters from XMM registers on Windows x64.

// goTR_DrawUnderline — no-op.
func goTR_DrawUnderline(this, clientDrawingContext, baselineOriginX, baselineOriginY, underline, clientDrawingEffect uintptr) uintptr {
	return _S_OK
}

// goTR_DrawStrikethrough — no-op.
func goTR_DrawStrikethrough(this, clientDrawingContext, baselineOriginX, baselineOriginY, strikethrough, clientDrawingEffect uintptr) uintptr {
	return _S_OK
}

// goTR_DrawInlineObject — no-op.
func goTR_DrawInlineObject(this, clientDrawingContext, originX, originY, inlineObject uintptr, isSideways, isRightToLeft uintptr, clientDrawingEffect uintptr) uintptr {
	return _S_OK
}

// --- gdiInteropBackend methods ---

func (b *gdiInteropBackend) Init(factory uintptr) error {
	// IDWriteFactory::GetGdiInterop (vtable index 17)
	var gdiInterop uintptr
	_, err := comCall(factory, 17, uintptr(unsafe.Pointer(&gdiInterop)))
	if err != nil {
		return err
	}
	b.gdiInterop = gdiInterop

	// IDWriteFactory::CreateRenderingParams (vtable index 10)
	var renderParams uintptr
	_, err = comCall(factory, 10, uintptr(unsafe.Pointer(&renderParams)))
	if err != nil {
		comRelease(gdiInterop)
		b.gdiInterop = 0
		return err
	}
	b.renderParams = renderParams

	// Create initial bitmap target (small, will be resized)
	err = b.createBitmapTarget(1, 1)
	if err != nil {
		comRelease(renderParams)
		comRelease(gdiInterop)
		b.renderParams = 0
		b.gdiInterop = 0
		return err
	}

	// Create the Go text renderer COM object.
	vtable := initGoTextRendererVtable()
	b.renderer = &goTextRenderer{
		vtable:       vtable,
		refCount:     1,
		bitmapTarget: b.bitmapTarget,
		renderParams: b.renderParams,
	}

	return nil
}

func (b *gdiInteropBackend) createBitmapTarget(width, height int) error {
	if b.bitmapTarget != 0 {
		comRelease(b.bitmapTarget)
		b.bitmapTarget = 0
	}

	// IDWriteGdiInterop::CreateBitmapRenderTarget (vtable index 7)
	// Signature: HRESULT CreateBitmapRenderTarget(HDC hdc, UINT32 width, UINT32 height,
	//   IDWriteBitmapRenderTarget **renderTarget)
	var target uintptr
	_, err := comCall(b.gdiInterop, 7,
		0, // hdc = NULL (use screen DC)
		uintptr(uint32(width)),
		uintptr(uint32(height)),
		uintptr(unsafe.Pointer(&target)),
	)
	if err != nil {
		return err
	}
	b.bitmapTarget = target
	b.width = width
	b.height = height

	// Force 1:1 pixel mapping — the node tree already scales font sizes
	// from dp to physical pixels via ScaleNodeDPI, so we must prevent
	// IDWriteBitmapRenderTarget from applying its own DPI scaling.
	// IDWriteBitmapRenderTarget::SetPixelsPerDip (vtable index 6)
	comCall(target, 6, uintptr(math.Float32bits(1.0)))

	return nil
}

func (b *gdiInteropBackend) BeginDraw(width, height int) error {
	// Always recreate bitmap target at exact size needed.
	// (Resize can silently fail or leave internal state inconsistent.)
	if width != b.width || height != b.height || b.bitmapTarget == 0 {
		if err := b.createBitmapTarget(width, height); err != nil {
			return err
		}
		// Update renderer's bitmapTarget pointer.
		if b.renderer != nil {
			b.renderer.bitmapTarget = b.bitmapTarget
		}
	}

	// Clear the bitmap to black using direct memory zeroing.
	// IDWriteBitmapRenderTarget::GetMemoryDC (vtable index 4)
	// Signature: HDC GetMemoryDC() — returns HDC directly (not HRESULT).
	memDC, _ := comCall(b.bitmapTarget, 4)
	if memDC != 0 {
		// Get the bitmap from the DC and clear it via GetDIBits/SetDIBits
		hBitmap, _, _ := procGetCurrentObject.Call(memDC, _OBJ_BITMAP)
		if hBitmap != 0 {
			bmi := BITMAPINFO{
				BmiHeader: BITMAPINFOHEADER{
					BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
					BiWidth:       int32(width),
					BiHeight:      -int32(height),
					BiPlanes:      1,
					BiBitCount:    32,
					BiCompression: BI_RGB,
				},
			}
			clearBuf := make([]byte, width*height*4) // all zeros = black
			procSetDIBits.Call(
				memDC, hBitmap,
				0, uintptr(height),
				uintptr(unsafe.Pointer(&clearBuf[0])),
				uintptr(unsafe.Pointer(&bmi)),
				DIB_RGB_COLORS,
			)
		}
	}
	return nil
}

func (b *gdiInteropBackend) DrawGlyphRun(baselineOriginX, baselineOriginY float32,
	measuringMode uint32, glyphRun uintptr, glyphRunDesc uintptr, textColor uint32) error {

	if b.bitmapTarget == 0 {
		return syscall.EINVAL
	}

	var blackBoxRect RECT
	_, err := comCall(b.bitmapTarget, 3,
		uintptr(math.Float32bits(baselineOriginX)),
		uintptr(math.Float32bits(baselineOriginY)),
		uintptr(measuringMode),
		glyphRun,
		b.renderParams,
		uintptr(textColor),
		uintptr(unsafe.Pointer(&blackBoxRect)),
	)
	return err
}

func (b *gdiInteropBackend) EndDraw() ([]byte, int, error) {
	if b.bitmapTarget == 0 {
		return nil, 0, syscall.EINVAL
	}

	// Get memory DC.
	memDC, _ := comCall(b.bitmapTarget, 4)
	if memDC == 0 {
		return nil, 0, syscall.EINVAL
	}

	// Get the current bitmap from the DC.
	hBitmap, _, _ := procGetCurrentObject.Call(memDC, _OBJ_BITMAP)
	if hBitmap == 0 {
		return nil, 0, syscall.EINVAL
	}

	// Prepare BITMAPINFO for GetDIBits — request top-down 32bpp BGRA.
	bmi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       int32(b.width),
			BiHeight:      -int32(b.height), // negative = top-down
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}

	stride := b.width * 4
	pixels := make([]byte, stride*b.height)

	ret, _, _ := procGetDIBits.Call(
		memDC,
		hBitmap,
		0,
		uintptr(b.height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, 0, syscall.EINVAL
	}

	return pixels, stride, nil
}

func (b *gdiInteropBackend) Close() {
	if b.bitmapTarget != 0 {
		comRelease(b.bitmapTarget)
		b.bitmapTarget = 0
	}
	if b.renderParams != 0 {
		comRelease(b.renderParams)
		b.renderParams = 0
	}
	if b.gdiInterop != 0 {
		comRelease(b.gdiInterop)
		b.gdiInterop = 0
	}
	b.renderer = nil
}

// DWriteTextRenderer implements core.TextRenderer using DirectWrite COM APIs.
type DWriteTextRenderer struct {
	factory    uintptr // IDWriteFactory*
	textFormat uintptr // IDWriteTextFormat* (current font)

	fontFamily string
	fontWeight int
	fontSize   float64

	// GDI Interop backend for pixel rendering.
	backend *gdiInteropBackend

	mu sync.Mutex
}

// NewDWriteTextRenderer creates a DirectWrite-based text renderer.
func NewDWriteTextRenderer() (*DWriteTextRenderer, error) {
	if err := initDWrite(); err != nil {
		return nil, err
	}

	var factory uintptr
	hr, _, _ := procDWriteCreateFactory.Call(
		DWRITE_FACTORY_TYPE_SHARED,
		uintptr(unsafe.Pointer(&iidIDWriteFactory)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if int32(hr) < 0 || factory == 0 {
		return nil, syscall.Errno(hr)
	}

	tr := &DWriteTextRenderer{
		factory:    factory,
		fontFamily: "Microsoft YaHei UI",
		fontWeight: DWRITE_FONT_WEIGHT_NORMAL,
		fontSize:   14,
	}

	// Create initial text format
	if err := tr.recreateTextFormat(); err != nil {
		comRelease(factory)
		return nil, err
	}

	return tr, nil
}

// recreateTextFormat creates a new IDWriteTextFormat with current font settings.
// IDWriteFactory::CreateTextFormat is at vtable index 15.
func (tr *DWriteTextRenderer) recreateTextFormat() error {
	if tr.textFormat != 0 {
		comRelease(tr.textFormat)
		tr.textFormat = 0
	}

	familyName, _ := syscall.UTF16PtrFromString(tr.fontFamily)
	localeName, _ := syscall.UTF16PtrFromString("en-us")

	var textFormat uintptr
	// IDWriteFactory vtable:
	// 0: QueryInterface, 1: AddRef, 2: Release
	// 3: GetSystemFontCollection, 4: CreateCustomFontCollection
	// 5: RegisterFontCollectionLoader, 6: UnregisterFontCollectionLoader
	// 7: CreateFontFileReference, 8: CreateCustomFontFileReference
	// 9: CreateFontFace, 10: CreateRenderingParams
	// 11: CreateMonitorRenderingParams, 12: CreateCustomRenderingParams
	// 13: RegisterFontFileLoader, 14: UnregisterFontFileLoader
	// 15: CreateTextFormat
	_, err := comCall(tr.factory, 15,
		uintptr(unsafe.Pointer(familyName)),                  // fontFamilyName
		0,                                                     // fontCollection (NULL = system)
		uintptr(tr.fontWeight),                                // fontWeight
		DWRITE_FONT_STYLE_NORMAL,                              // fontStyle
		DWRITE_FONT_STRETCH_NORMAL,                            // fontStretch
		uintptr(math.Float32bits(float32(tr.fontSize))),       // fontSize
		uintptr(unsafe.Pointer(localeName)),                   // localeName
		uintptr(unsafe.Pointer(&textFormat)),                  // out: textFormat
	)
	if err != nil {
		return err
	}
	tr.textFormat = textFormat
	return nil
}

// createTextLayout creates an IDWriteTextLayout for the given text.
// IDWriteFactory::CreateTextLayout is at vtable index 18.
func (tr *DWriteTextRenderer) createTextLayout(text string, maxWidth, maxHeight float64) (uintptr, error) {
	if tr.textFormat == 0 {
		return 0, syscall.EINVAL
	}

	textUTF16, _ := syscall.UTF16FromString(text)
	var layout uintptr

	// IDWriteFactory vtable:
	// 16: CreateTypography, 17: GetGdiInterop
	// 18: CreateTextLayout
	_, err := comCall(tr.factory, 18,
		uintptr(unsafe.Pointer(&textUTF16[0])),       // string
		uintptr(uint32(len(textUTF16)-1)),             // stringLength (exclude null terminator)
		tr.textFormat,                                  // textFormat
		uintptr(math.Float32bits(float32(maxWidth))),  // maxWidth
		uintptr(math.Float32bits(float32(maxHeight))), // maxHeight
		uintptr(unsafe.Pointer(&layout)),               // out: textLayout
	)
	if err != nil {
		return 0, err
	}
	return layout, nil
}

// getTextMetrics calls IDWriteTextLayout::GetMetrics.
// IDWriteTextLayout extends IDWriteTextFormat (28 methods: 0-27).
// IDWriteTextLayout own methods start at 28:
// 28: SetMaxWidth, 29: SetMaxHeight, 30: SetFontCollection
// 31: SetFontFamilyName, 32: SetFontWeight, 33: SetFontStyle
// 34: SetFontStretch, 35: SetFontSize, 36: SetUnderline
// 37: SetStrikethrough, 38: SetDrawingEffect, 39: SetInlineObject
// 40: SetTypography, 41: SetLocaleName
// 42: GetMaxWidth, 43: GetMaxHeight
// 44: GetFontCollection, 45: GetFontFamilyNameLength, 46: GetFontFamilyName
// 47: GetFontWeight, 48: GetFontStyle, 49: GetFontStretch
// 50: GetFontSize, 51: GetUnderline, 52: GetStrikethrough
// 53: GetDrawingEffect, 54: GetInlineObject, 55: GetTypography
// 56: GetLocaleNameLength, 57: GetLocaleName
// 58: Draw
// 59: GetLineMetrics
// 60: GetMetrics
func getTextMetrics(layout uintptr) (dwriteTextMetrics, error) {
	var metrics dwriteTextMetrics
	_, err := comCall(layout, 60,
		uintptr(unsafe.Pointer(&metrics)),
	)
	return metrics, err
}

// getLineMetrics calls IDWriteTextLayout::GetLineMetrics.
func getLineMetrics(layout uintptr) ([]dwriteLineMetrics, error) {
	// First call to get count
	var lineCount uint32
	comCall(layout, 59, 0, 0, uintptr(unsafe.Pointer(&lineCount)))

	if lineCount == 0 {
		return nil, nil
	}

	lines := make([]dwriteLineMetrics, lineCount)
	var actualCount uint32
	_, err := comCall(layout, 59,
		uintptr(unsafe.Pointer(&lines[0])),
		uintptr(lineCount),
		uintptr(unsafe.Pointer(&actualCount)),
	)
	if err != nil {
		return nil, err
	}
	return lines[:actualCount], nil
}

// --- core.TextRenderer interface ---

func (tr *DWriteTextRenderer) SetFont(fontFamily string, weight int, size float64) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	changed := false
	if fontFamily != "" && fontFamily != tr.fontFamily {
		tr.fontFamily = fontFamily
		changed = true
	}
	if weight > 0 && weight != tr.fontWeight {
		tr.fontWeight = weight
		changed = true
	}
	if size > 0 && size != tr.fontSize {
		tr.fontSize = size
		changed = true
	}
	if changed {
		tr.recreateTextFormat()
	}
}

func (tr *DWriteTextRenderer) MeasureText(text string) core.Size {
	if text == "" {
		return core.Size{}
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	layout, err := tr.createTextLayout(text, 100000, 100000) // huge max for single-line measure
	if err != nil || layout == 0 {
		return core.Size{}
	}
	defer comRelease(layout)

	metrics, err := getTextMetrics(layout)
	if err != nil {
		return core.Size{}
	}

	return core.Size{
		Width:  float64(metrics.WidthIncludingTrailingWhitespace),
		Height: float64(metrics.Height),
	}
}

func (tr *DWriteTextRenderer) DrawText(canvas core.Canvas, text string, x, y float64, paint *core.Paint) {
	if text == "" || canvas == nil {
		return
	}
	target := canvas.Target()
	if target == nil {
		return
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Lazily initialize GDI Interop backend (for bitmap target DC).
	if tr.backend == nil {
		b := &gdiInteropBackend{}
		if err := b.Init(tr.factory); err != nil {
			return
		}
		tr.backend = b
	}

	// Determine text color as COLORREF (0x00BBGGRR).
	var clrR, clrG, clrB uint8
	if paint != nil && paint.Color.A != 0 {
		clrR = paint.Color.R
		clrG = paint.Color.G
		clrB = paint.Color.B
	}
	colorRef := uint32(clrB)<<16 | uint32(clrG)<<8 | uint32(clrR)

	// Measure text to determine bitmap size (DirectWrite measurement).
	size := tr.measureTextLocked(text)
	bmpW := int(math.Ceil(size.Width)) + 4
	bmpH := int(math.Ceil(size.Height)) + 4
	if bmpW <= 0 || bmpH <= 0 {
		return
	}

	// Prepare bitmap (clear to black).
	if err := tr.backend.BeginDraw(bmpW, bmpH); err != nil {
		return
	}

	// Always render WHITE text on the black bitmap to get a coverage mask.
	// compositeToCanvas uses max(R,G,B) as alpha and applies the actual
	// text color (clrR, clrG, clrB) during compositing.
	_ = colorRef
	tr.backend.renderer.textColor = 0x00FFFFFF

	// Create a text layout for rendering via DirectWrite.
	layout, err := tr.createTextLayout(text, float64(bmpW), float64(bmpH))
	if err != nil || layout == 0 {
		return
	}
	defer comRelease(layout)

	// Get baseline from line metrics for correct vertical positioning.
	lineMetricsList, _ := getLineMetrics(layout)
	var baselineY float32
	if len(lineMetricsList) > 0 {
		baselineY = lineMetricsList[0].Baseline
	}

	// Render text using IDWriteTextLayout::Draw (vtable index 58).
	// This calls our goTextRenderer COM callback chain:
	//   Draw() → DrawGlyphRun() (CGO trampoline) → IDWriteBitmapRenderTarget::DrawGlyphRun
	// The renderer pointer is passed as the custom text renderer.
	comCall(layout, 58,
		0, // clientDrawingContext (unused)
		uintptr(unsafe.Pointer(tr.backend.renderer)), // IDWriteTextRenderer*
		uintptr(math.Float32bits(0)),                  // originX
		uintptr(math.Float32bits(0)),                  // originY
	)
	_ = baselineY // baseline handled by DirectWrite layout internally

	// Read pixels from the DirectWrite bitmap.
	pixels, stride, err := tr.backend.EndDraw()
	if err != nil {
		return
	}

	// Composite BGRA pixels onto the canvas RGBA image.
	// DirectWrite renders text on black background via IDWriteBitmapRenderTarget.
	// Use max(R,G,B) as coverage alpha, apply requested text color.
	compositeToCanvas(target, int(x), int(y), bmpW, bmpH, pixels, stride, clrR, clrG, clrB)
}

// measureTextLocked measures text — must be called with tr.mu held.
func (tr *DWriteTextRenderer) measureTextLocked(text string) core.Size {
	layout, err := tr.createTextLayout(text, 100000, 100000)
	if err != nil || layout == 0 {
		return core.Size{}
	}
	defer comRelease(layout)

	metrics, err := getTextMetrics(layout)
	if err != nil {
		return core.Size{}
	}
	return core.Size{
		Width:  float64(metrics.WidthIncludingTrailingWhitespace),
		Height: float64(metrics.Height),
	}
}

// compositeToCanvas blits BGRA pixel data from the DirectWrite bitmap onto
// the target image.RGBA at position (dx, dy).
//
// IDWriteBitmapRenderTarget renders text on a black background. The RGB channels
// contain the text color blended with black, and we use the luminance as the
// coverage (alpha) value. For colored text we reconstruct alpha from the
// maximum of the RGB channels and use the requested text color.
func compositeToCanvas(target *image.RGBA, dx, dy, srcW, srcH int, pixels []byte, stride int, textR, textG, textB uint8) {
	bounds := target.Bounds()
	for py := 0; py < srcH; py++ {
		dstY := dy + py
		if dstY < bounds.Min.Y || dstY >= bounds.Max.Y {
			continue
		}
		for px := 0; px < srcW; px++ {
			dstX := dx + px
			if dstX < bounds.Min.X || dstX >= bounds.Max.X {
				continue
			}

			si := py*stride + px*4
			if si+3 >= len(pixels) {
				continue
			}

			// BGRA layout from GetDIBits.
			sb := pixels[si+0]
			sg := pixels[si+1]
			sr := pixels[si+2]
			// pixels[si+3] is always 0 for GDI bitmaps (no alpha channel).

			// Use maximum channel as coverage — handles both grayscale and ClearType AA.
			alpha := sr
			if sg > alpha {
				alpha = sg
			}
			if sb > alpha {
				alpha = sb
			}
			if alpha == 0 {
				continue // fully transparent, skip
			}

			// Alpha-blend onto destination using "over" compositing.
			di := (dstY-bounds.Min.Y)*target.Stride + (dstX-bounds.Min.X)*4
			if di+3 >= len(target.Pix) {
				continue
			}

			a := uint32(alpha)
			invA := 255 - a

			dstR := target.Pix[di+0]
			dstG := target.Pix[di+1]
			dstB := target.Pix[di+2]
			dstA := target.Pix[di+3]

			target.Pix[di+0] = uint8((a*uint32(textR) + invA*uint32(dstR)) / 255)
			target.Pix[di+1] = uint8((a*uint32(textG) + invA*uint32(dstG)) / 255)
			target.Pix[di+2] = uint8((a*uint32(textB) + invA*uint32(dstB)) / 255)
			target.Pix[di+3] = uint8((a*255 + invA*uint32(dstA)) / 255)
		}
	}
}

func (tr *DWriteTextRenderer) CreateTextLayout(text string, paint *core.Paint, maxWidth float64) *core.TextLayoutResult {
	if text == "" {
		return &core.TextLayoutResult{}
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Apply paint font settings temporarily if different
	if paint != nil && paint.FontSize > 0 && paint.FontSize != tr.fontSize {
		oldSize := tr.fontSize
		tr.fontSize = paint.FontSize
		tr.recreateTextFormat()
		defer func() {
			tr.fontSize = oldSize
			tr.recreateTextFormat()
		}()
	}

	layout, err := tr.createTextLayout(text, maxWidth, 100000)
	if err != nil || layout == 0 {
		return &core.TextLayoutResult{}
	}
	defer comRelease(layout)

	// Get overall metrics
	textMetrics, _ := getTextMetrics(layout)

	// Get per-line metrics
	lineMetricsList, _ := getLineMetrics(layout)

	var lines []core.TextLine
	var currentY float64
	textRunes := []rune(text)
	runeIdx := 0

	for _, lm := range lineMetricsList {
		// Extract line text using Length (character count)
		lineLen := int(lm.Length - lm.NewlineLength)
		endIdx := runeIdx + lineLen
		if endIdx > len(textRunes) {
			endIdx = len(textRunes)
		}
		lineText := string(textRunes[runeIdx:endIdx])

		lines = append(lines, core.TextLine{
			Text:     lineText,
			Offset:   core.Point{X: 0, Y: currentY},
			Width:    0, // will be fixed below
			Baseline: currentY + float64(lm.Baseline),
		})
		currentY += float64(lm.Height)
		runeIdx += int(lm.Length)
	}

	// Fix line widths using per-line measurement
	for i := range lines {
		lineLayout, err := tr.createTextLayout(lines[i].Text, 100000, 100000)
		if err == nil && lineLayout != 0 {
			m, _ := getTextMetrics(lineLayout)
			lines[i].Width = float64(m.WidthIncludingTrailingWhitespace)
			comRelease(lineLayout)
		}
	}

	return &core.TextLayoutResult{
		Lines: lines,
		TotalSize: core.Size{
			Width:  float64(textMetrics.WidthIncludingTrailingWhitespace),
			Height: float64(textMetrics.Height),
		},
	}
}

func (tr *DWriteTextRenderer) Close() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.backend != nil {
		tr.backend.Close()
		tr.backend = nil
	}
	if tr.textFormat != 0 {
		comRelease(tr.textFormat)
		tr.textFormat = 0
	}
	if tr.factory != 0 {
		comRelease(tr.factory)
		tr.factory = 0
	}
}
