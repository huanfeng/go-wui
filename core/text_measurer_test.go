package core

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

// mockTextMeasurer is a fake TextMeasurer that returns predictable sizes.
type mockTextMeasurer struct {
	callCount int
}

func (m *mockTextMeasurer) MeasureText(text string, paint *Paint) Size {
	m.callCount++
	if text == "" {
		return Size{}
	}
	fontSize := 14.0
	if paint != nil && paint.FontSize > 0 {
		fontSize = paint.FontSize
	}
	// Use a simple formula: each rune = fontSize * 0.8 wide
	runeCount := float64(len([]rune(text)))
	return Size{Width: runeCount * fontSize * 0.8, Height: fontSize * 1.2}
}

func TestNodeMeasureText_WithTextMeasurer(t *testing.T) {
	root := NewNode("root")
	child := NewNode("child")
	root.AddChild(child)

	mock := &mockTextMeasurer{}
	root.SetData("textMeasurer", TextMeasurer(mock))

	paint := &Paint{FontSize: 20}
	size := NodeMeasureText(child, "Hello", paint)

	if mock.callCount != 1 {
		t.Errorf("expected 1 call to MeasureText, got %d", mock.callCount)
	}
	// "Hello" = 5 runes, fontSize=20 → width = 5*20*0.8 = 80, height = 20*1.2 = 24
	if size.Width != 80 {
		t.Errorf("expected width 80, got %v", size.Width)
	}
	if size.Height != 24 {
		t.Errorf("expected height 24, got %v", size.Height)
	}
}

func TestNodeMeasureText_FallbackWithoutTextMeasurer(t *testing.T) {
	node := NewNode("orphan")
	paint := &Paint{FontSize: 20}
	size := NodeMeasureText(node, "Hello", paint)

	// Fallback: 5 runes * 20 * 0.6 = 60, height = 20 * 1.4 = 28
	if size.Width != 60 {
		t.Errorf("expected fallback width 60, got %v", size.Width)
	}
	if size.Height != 28 {
		t.Errorf("expected fallback height 28, got %v", size.Height)
	}
}

func TestNodeMeasureText_EmptyString(t *testing.T) {
	node := NewNode("test")
	size := NodeMeasureText(node, "", &Paint{FontSize: 20})
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("empty text should return zero size, got %v", size)
	}
}

func TestGetTextMeasurer_WalksUpTree(t *testing.T) {
	root := NewNode("root")
	mid := NewNode("mid")
	leaf := NewNode("leaf")
	root.AddChild(mid)
	mid.AddChild(leaf)

	mock := &mockTextMeasurer{}
	root.SetData("textMeasurer", TextMeasurer(mock))

	tm := GetTextMeasurer(leaf)
	if tm == nil {
		t.Fatal("expected to find TextMeasurer from leaf via root")
	}
	if tm != TextMeasurer(mock) {
		t.Error("found wrong TextMeasurer")
	}
}

func TestGetTextMeasurer_ReturnsNilWhenMissing(t *testing.T) {
	node := NewNode("orphan")
	if GetTextMeasurer(node) != nil {
		t.Error("expected nil when no TextMeasurer is set")
	}
}

func TestNewTextMeasurer_WrapsTextRenderer(t *testing.T) {
	// Verify the adapter satisfies the interface
	var _ TextMeasurer = NewTextMeasurer(nil)
}

func TestNodeMeasureText_CJKCharacters(t *testing.T) {
	root := NewNode("root")
	mock := &mockTextMeasurer{}
	root.SetData("textMeasurer", TextMeasurer(mock))

	paint := &Paint{FontSize: 14}
	size := NodeMeasureText(root, "你好世界", paint)

	// 4 CJK runes, fontSize=14 → width = 4*14*0.8 = 44.8, height = 14*1.2 = 16.8
	if !approxEqual(size.Width, 44.8) {
		t.Errorf("expected CJK width ~44.8, got %v", size.Width)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
}

func TestNodeMeasureText_MixedContent(t *testing.T) {
	root := NewNode("root")
	mock := &mockTextMeasurer{}
	root.SetData("textMeasurer", TextMeasurer(mock))

	paint := &Paint{FontSize: 16}
	// "Hello你好" = 7 runes
	size := NodeMeasureText(root, "Hello你好", paint)

	expected := 7.0 * 16.0 * 0.8 // = 89.6
	if !approxEqual(size.Width, expected) {
		t.Errorf("expected mixed width ~%v, got %v", expected, size.Width)
	}
}
