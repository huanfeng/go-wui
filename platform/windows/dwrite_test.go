//go:build windows

package windows

import (
	"image/color"
	"syscall"
	"testing"
	"unsafe"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/render/gg"
)

func TestDWriteInit(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Fatalf("failed to create DWrite renderer: %v", err)
	}
	defer tr.Close()
}

func TestDWriteMeasureASCII(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	size := tr.MeasureText("Hello World")
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("expected positive size, got %+v", size)
	}
	t.Logf("'Hello World' = %.1f x %.1f", size.Width, size.Height)
}

func TestDWriteMeasureCJK(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	tr.SetFont("Microsoft YaHei", 400, 16)
	size := tr.MeasureText("你好世界")
	if size.Width <= 0 {
		t.Error("CJK text should have positive width")
	}
	t.Logf("'你好世界' = %.1f x %.1f", size.Width, size.Height)
}

func TestDWriteMeasureEmpty(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	size := tr.MeasureText("")
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("empty text should be zero size, got %+v", size)
	}
}

func TestDWriteLongerIsWider(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	short := tr.MeasureText("Hi")
	long := tr.MeasureText("Hello World")
	if long.Width <= short.Width {
		t.Errorf("longer text should be wider: short=%+v long=%+v", short, long)
	}
}

func TestDWriteCreateTextLayout(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	paint := &core.Paint{FontSize: 14}
	result := tr.CreateTextLayout("Hello World this is a long text for testing line breaks", paint, 100)
	if len(result.Lines) < 2 {
		t.Errorf("expected multiple lines for narrow width, got %d", len(result.Lines))
	}
	for i, line := range result.Lines {
		t.Logf("Line %d: %q (width=%.1f, baseline=%.1f)", i, line.Text, line.Width, line.Baseline)
	}
}

func TestDWriteFontSizeChange(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	tr.SetFont("", 0, 12)
	small := tr.MeasureText("Hello")

	tr.SetFont("", 0, 24)
	large := tr.MeasureText("Hello")

	if large.Width <= small.Width || large.Height <= small.Height {
		t.Errorf("larger font should produce larger text: small=%+v large=%+v", small, large)
	}
}

func TestDWriteGdiInteropInit(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	// Initialize the backend directly.
	b := &gdiInteropBackend{}
	err = b.Init(tr.factory)
	if err != nil {
		t.Fatalf("GDI Interop backend Init failed: %v", err)
	}
	defer b.Close()

	if b.gdiInterop == 0 {
		t.Error("gdiInterop should be non-zero")
	}
	if b.bitmapTarget == 0 {
		t.Error("bitmapTarget should be non-zero")
	}
	if b.renderParams == 0 {
		t.Error("renderParams should be non-zero")
	}
	if b.renderer == nil {
		t.Error("renderer should be non-nil")
	}
}

func TestDWriteDrawTextProducesPixels(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	// Create a canvas big enough for the text.
	canvas := gg.NewGGCanvas(200, 50, tr)

	paint := &core.Paint{
		Color:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // white text
		FontSize: 14,
	}

	tr.SetFont("Segoe UI", 400, 14)
	tr.DrawText(canvas, "Hello", 10, 10, paint)

	// Check that at least some pixels were written.
	img := canvas.Target()
	nonZero := 0
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 || a > 0 {
				nonZero++
			}
		}
	}

	t.Logf("DrawText produced %d non-zero pixels out of %d total",
		nonZero, bounds.Dx()*bounds.Dy())

	if nonZero == 0 {
		t.Error("DrawText should produce non-zero pixels for visible text")
	}
}

func TestDWriteDrawTextBlackOnWhite(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	// Create a canvas and fill with white background.
	canvas := gg.NewGGCanvas(200, 50, tr)
	img := canvas.Target()
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = 255 // R
		img.Pix[i+1] = 255 // G
		img.Pix[i+2] = 255 // B
		img.Pix[i+3] = 255 // A
	}

	paint := &core.Paint{
		Color:    color.RGBA{R: 0, G: 0, B: 0, A: 255}, // black text
		FontSize: 14,
	}

	tr.SetFont("Segoe UI", 400, 14)
	tr.DrawText(canvas, "Test", 10, 10, paint)

	// Check that some pixels are darker than pure white (text was rendered).
	darkPixels := 0
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Pixel is "dark" if any channel is below full white.
			if r < 0xFFFF || g < 0xFFFF || b < 0xFFFF {
				darkPixels++
			}
		}
	}

	t.Logf("DrawText produced %d dark pixels on white background", darkPixels)

	if darkPixels == 0 {
		t.Error("DrawText with black color on white background should produce dark pixels")
	}
}

// Diagnostic: test bitmap target DC directly with GDI TextOut
func TestDWriteDiag_GDITextOnBitmapDC(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Initialize backend
	if tr.backend == nil {
		b := &gdiInteropBackend{}
		if err := b.Init(tr.factory); err != nil {
			t.Fatalf("backend init failed: %v", err)
		}
		tr.backend = b
	}

	// Resize bitmap target to 100x30
	w, h := 100, 30
	if err := tr.backend.BeginDraw(w, h); err != nil {
		t.Fatalf("BeginDraw failed: %v", err)
	}

	// Get memory DC
	memDC, _ := comCall(tr.backend.bitmapTarget, 4)
	t.Logf("memDC = %v", memDC)
	if memDC == 0 {
		t.Fatal("GetMemoryDC returned 0")
	}

	// Draw text directly with GDI TextOutW
	textOutW := syscall.NewLazyDLL("gdi32.dll").NewProc("TextOutW")
	setBkMode := syscall.NewLazyDLL("gdi32.dll").NewProc("SetBkMode")
	setTextColor := syscall.NewLazyDLL("gdi32.dll").NewProc("SetTextColor")

	setBkMode.Call(memDC, 1) // TRANSPARENT = 1
	setTextColor.Call(memDC, 0x00FFFFFF) // white

	text, _ := syscall.UTF16FromString("Test")
	textOutW.Call(memDC, 5, 5, uintptr(unsafe.Pointer(&text[0])), uintptr(len(text)-1))

	// Read pixels back
	pixels, stride, err := tr.backend.EndDraw()
	if err != nil {
		t.Fatalf("EndDraw failed: %v", err)
	}

	nonZero := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := y*stride + x*4
			if si+2 < len(pixels) {
				if pixels[si] > 0 || pixels[si+1] > 0 || pixels[si+2] > 0 {
					nonZero++
				}
			}
		}
	}
	t.Logf("GDI TextOut produced %d non-zero pixels (stride=%d, len=%d)", nonZero, stride, len(pixels))
	if nonZero == 0 {
		t.Error("GDI TextOut should produce non-zero pixels — DC or GetDIBits is broken")
	}
}

func TestDWriteDrawTextEmptyNoPanic(t *testing.T) {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		t.Skip("DirectWrite not available:", err)
	}
	defer tr.Close()

	canvas := gg.NewGGCanvas(100, 30, tr)
	paint := &core.Paint{
		Color:    color.RGBA{R: 255, A: 255},
		FontSize: 14,
	}

	// These should be no-ops, not panics.
	tr.DrawText(canvas, "", 0, 0, paint)
	tr.DrawText(nil, "Hello", 0, 0, paint)
}
