//go:build windows

package windows

import (
	"math"
	"sync"
	"syscall"
	"unsafe"

	"gowui/core"
	"gowui/render/freetype"
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

// DLL and proc
var (
	dwriteDll               *syscall.LazyDLL
	procDWriteCreateFactory *syscall.LazyProc
	dwriteInitOnce          sync.Once
	dwriteInitErr           error
)

func initDWrite() error {
	dwriteInitOnce.Do(func() {
		dwriteDll = syscall.NewLazyDLL("dwrite.dll")
		procDWriteCreateFactory = dwriteDll.NewProc("DWriteCreateFactory")
		dwriteInitErr = procDWriteCreateFactory.Find()
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

// DWriteTextRenderer implements core.TextRenderer using DirectWrite COM APIs.
type DWriteTextRenderer struct {
	factory    uintptr // IDWriteFactory*
	textFormat uintptr // IDWriteTextFormat* (current font)

	fontFamily string
	fontWeight int
	fontSize   float64

	// FreeType fallback for pixel rendering (until D2D is integrated)
	drawFallback *freetype.FreeTypeTextRenderer

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
		fontFamily: "Segoe UI",
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
	// Use FreeType for pixel rendering until D2D integration in Phase 2.
	// Measurement is handled by DirectWrite (MeasureText/CreateTextLayout).
	if tr.drawFallback == nil {
		tr.drawFallback = freetype.NewFreeTypeTextRenderer()
	}
	tr.drawFallback.DrawText(canvas, text, x, y, paint)
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

	if tr.textFormat != 0 {
		comRelease(tr.textFormat)
		tr.textFormat = 0
	}
	if tr.factory != 0 {
		comRelease(tr.factory)
		tr.factory = 0
	}
}
