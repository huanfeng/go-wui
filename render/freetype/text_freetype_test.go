package freetype

import (
	"testing"

	"github.com/huanfeng/wind-ui/core"
)

func TestNewFreeTypeTextRenderer(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	if tr.face == nil {
		t.Fatal("face should not be nil")
	}
	if tr.fontSize != 13 {
		t.Errorf("default fontSize should be 13, got %v", tr.fontSize)
	}
}

func TestMeasureText_NonZero(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	size := tr.MeasureText("Hello")
	if size.Width <= 0 {
		t.Errorf("width should be > 0, got %v", size.Width)
	}
	if size.Height <= 0 {
		t.Errorf("height should be > 0, got %v", size.Height)
	}
}

func TestMeasureText_Empty(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	size := tr.MeasureText("")
	if size.Width != 0 {
		t.Error("empty string should have zero width")
	}
	if size.Height != 0 {
		t.Error("empty string should have zero height")
	}
}

func TestMeasureText_LongerIsWider(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	short := tr.MeasureText("Hi")
	long := tr.MeasureText("Hello World")
	if long.Width <= short.Width {
		t.Errorf("longer text should be wider: short=%v, long=%v", short.Width, long.Width)
	}
}

func TestMeasureText_ScalesWithFontSize(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()

	base := tr.MeasureText("Test")

	tr.SetFont("", 0, 26) // double the default 13
	scaled := tr.MeasureText("Test")

	if scaled.Width <= base.Width {
		t.Errorf("scaled width (%v) should be larger than base width (%v)", scaled.Width, base.Width)
	}
	if scaled.Height <= base.Height {
		t.Errorf("scaled height (%v) should be larger than base height (%v)", scaled.Height, base.Height)
	}
}

func TestSetFont_PositiveSize(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	tr.SetFont("Arial", 400, 20)
	if tr.fontSize != 20 {
		t.Errorf("fontSize should be 20, got %v", tr.fontSize)
	}
}

func TestSetFont_ZeroSizeIgnored(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	tr.SetFont("Arial", 400, 0)
	if tr.fontSize != 13 {
		t.Errorf("fontSize should remain 13 when zero is passed, got %v", tr.fontSize)
	}
}

func TestCreateTextLayout_SingleLine(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	result := tr.CreateTextLayout("Hello", paint, 1000) // wide enough for one line
	if len(result.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(result.Lines))
	}
	if result.Lines[0].Text != "Hello" {
		t.Errorf("line text mismatch: got %q", result.Lines[0].Text)
	}
	if result.TotalSize.Width <= 0 {
		t.Error("total width should be > 0")
	}
	if result.TotalSize.Height <= 0 {
		t.Error("total height should be > 0")
	}
}

func TestCreateTextLayout_LineBreak(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	// Force line break with narrow width.
	result := tr.CreateTextLayout("Hello World Test", paint, 50)
	if len(result.Lines) < 2 {
		t.Errorf("expected multiple lines with narrow width, got %d", len(result.Lines))
	}
	// Each line should have some text.
	for i, line := range result.Lines {
		if line.Text == "" {
			t.Errorf("line %d should not be empty", i)
		}
	}
}

func TestCreateTextLayout_Empty(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	result := tr.CreateTextLayout("", paint, 100)
	if len(result.Lines) != 0 {
		t.Errorf("empty text should have no lines, got %d", len(result.Lines))
	}
}

func TestCreateTextLayout_ZeroMaxWidth(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	// maxWidth=0 means no wrapping.
	result := tr.CreateTextLayout("Hello World Test", paint, 0)
	if len(result.Lines) != 1 {
		t.Errorf("expected 1 line with maxWidth=0 (no wrapping), got %d", len(result.Lines))
	}
}

func TestCreateTextLayout_Baseline(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	result := tr.CreateTextLayout("Hello", paint, 1000)
	if len(result.Lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	if result.Lines[0].Baseline <= 0 {
		t.Errorf("baseline should be > 0, got %v", result.Lines[0].Baseline)
	}
}

func TestDrawText_NilCanvas(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	// Should not panic with nil canvas.
	tr.DrawText(nil, "Hello", 0, 0, paint)
}

func TestDrawText_EmptyString(t *testing.T) {
	tr := NewFreeTypeTextRenderer()
	defer tr.Close()
	paint := &core.Paint{FontSize: 13}
	// Should not panic with empty string.
	tr.DrawText(nil, "", 0, 0, paint)
}

// Verify the interface is satisfied at compile time.
var _ core.TextRenderer = (*FreeTypeTextRenderer)(nil)
