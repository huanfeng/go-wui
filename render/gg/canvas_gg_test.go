package gg

import (
	"image/color"
	"testing"

	"github.com/huanfeng/wind-ui/core"
)

// hasColor checks that at least one of R, G, B channels is non-zero and alpha
// is non-zero. With anti-aliasing, exact values may differ from the input
// color, so we just verify "something was drawn".
func hasColor(t *testing.T, c *GGCanvas, x, y int, desc string) {
	t.Helper()
	r, g, b, a := c.Target().At(x, y).RGBA()
	if (r == 0 && g == 0 && b == 0) || a == 0 {
		t.Errorf("%s: expected non-transparent colored pixel at (%d,%d), got RGBA(%d,%d,%d,%d)",
			desc, x, y, r, g, b, a)
	}
}

// isTransparent checks that the pixel is fully transparent (alpha == 0).
func isTransparent(t *testing.T, c *GGCanvas, x, y int, desc string) {
	t.Helper()
	_, _, _, a := c.Target().At(x, y).RGBA()
	if a != 0 {
		t.Errorf("%s: expected transparent pixel at (%d,%d), got alpha=%d", desc, x, y, a)
	}
}

func TestGGCanvasDrawRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 10, Y: 10, Width: 50, Height: 50}, paint)

	hasColor(t, c, 25, 25, "inside rect")
	isTransparent(t, c, 5, 5, "outside rect")
}

func TestGGCanvasDrawRectStroke(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{
		Color:       color.RGBA{R: 255, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 2,
	}
	c.DrawRect(core.Rect{X: 10, Y: 10, Width: 50, Height: 50}, paint)

	// On the top edge
	hasColor(t, c, 30, 10, "top stroke edge")

	// Interior should be empty
	isTransparent(t, c, 35, 35, "interior of stroked rect")
}

func TestGGCanvasSaveRestore(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Save()
	c.Translate(50, 50)
	c.Restore()

	paint := &core.Paint{Color: color.RGBA{G: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, paint)

	_, g, _, _ := c.Target().At(5, 5).RGBA()
	if g == 0 {
		t.Error("expected green pixel at original position after restore")
	}
}

func TestGGCanvasTranslate(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Translate(20, 20)
	paint := &core.Paint{Color: color.RGBA{B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, paint)

	// Rect should be at (20, 20) due to translate
	_, _, b, _ := c.Target().At(25, 25).RGBA()
	if b == 0 {
		t.Error("expected blue pixel at translated position (25,25)")
	}

	// Original position should be empty
	isTransparent(t, c, 5, 5, "original position after translate")
}

func TestGGCanvasClipRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.ClipRect(core.Rect{X: 20, Y: 20, Width: 30, Height: 30})
	paint := &core.Paint{Color: color.RGBA{R: 255, G: 255, A: 255}, DrawStyle: core.PaintFill}
	// Draw a rect that extends beyond the clip
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 100, Height: 100}, paint)

	// Inside clip: should have the color
	hasColor(t, c, 30, 30, "inside clip")

	// Outside clip: should be transparent
	isTransparent(t, c, 5, 5, "outside clip")
}

func TestGGCanvasDrawCircle(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 128, G: 128, B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawCircle(50, 50, 20, paint)

	// Center should be filled
	hasColor(t, c, 50, 50, "circle center")

	// Far outside the circle
	isTransparent(t, c, 5, 5, "outside circle")
}

func TestGGCanvasDrawLine(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintStroke}
	c.DrawLine(0, 0, 99, 99, paint)

	// Diagonal line: pixel near the middle
	hasColor(t, c, 50, 50, "diagonal line midpoint")
}

func TestGGCanvasDrawRoundRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{G: 200, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRoundRect(core.Rect{X: 10, Y: 10, Width: 60, Height: 60}, 10, paint)

	// Center should be filled
	hasColor(t, c, 40, 40, "center of round rect")

	// A corner pixel that is outside the rounded corner should be empty.
	// (10,10) is the exact corner — with radius 10, this pixel is outside the curve.
	isTransparent(t, c, 10, 10, "rounded corner")
}

func TestGGCanvasDrawImage(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)

	// Create a small red source image (10x10)
	srcImg := core.ImageResource{
		Width:  10,
		Height: 10,
		Name:   "test",
	}
	srcImg.Image = NewGGCanvas(10, 10, nil).Target()
	// Fill the source image with red
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			srcImg.Image.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	c.DrawImage(&srcImg, core.Rect{X: 20, Y: 20, Width: 10, Height: 10})

	hasColor(t, c, 25, 25, "drawn image pixel")
	isTransparent(t, c, 5, 5, "outside drawn image")
}

func TestGGCanvasNilPaintNoOp(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	// These should not panic
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, nil)
	c.DrawRoundRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, 5, nil)
	c.DrawCircle(50, 50, 20, nil)
	c.DrawLine(0, 0, 99, 99, nil)
	c.DrawText("hello", 10, 10, nil)
	_ = c.MeasureText("hello", nil)
}

func TestGGCanvasTranslateStacking(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Translate(10, 10)
	c.Translate(10, 10)
	// Total offset should be (20, 20)
	paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 5, Height: 5}, paint)

	hasColor(t, c, 22, 22, "after double translate")
	isTransparent(t, c, 5, 5, "original position after double translate")
}

func TestGGCanvasClipWithTranslate(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Translate(10, 10)
	c.ClipRect(core.Rect{X: 0, Y: 0, Width: 20, Height: 20})
	// Clip should be at (10, 10) to (30, 30) in image space
	paint := &core.Paint{Color: color.RGBA{B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: -100, Y: -100, Width: 500, Height: 500}, paint)

	// Inside clip
	hasColor(t, c, 20, 20, "inside clip with translate")

	// Outside clip
	isTransparent(t, c, 5, 5, "outside clip with translate")
}
