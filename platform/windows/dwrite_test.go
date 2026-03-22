//go:build windows

package windows

import (
	"testing"

	"gowui/core"
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
