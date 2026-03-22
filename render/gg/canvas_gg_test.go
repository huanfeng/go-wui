package gg

import (
	"image/color"
	"testing"

	"gowui/core"
)

func TestGGCanvasDrawRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 10, Y: 10, Width: 50, Height: 50}, paint)
	img := c.Target()

	// Inside the rect
	r, _, _, a := img.At(25, 25).RGBA()
	if r == 0 || a == 0 {
		t.Error("expected non-zero red pixel at (25,25)")
	}

	// Outside the rect
	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel outside rect at (5,5)")
	}
}

func TestGGCanvasDrawRectStroke(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{
		Color:       color.RGBA{R: 255, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 2,
	}
	c.DrawRect(core.Rect{X: 10, Y: 10, Width: 50, Height: 50}, paint)
	img := c.Target()

	// On the top edge
	r, _, _, a := img.At(30, 10).RGBA()
	if r == 0 || a == 0 {
		t.Error("expected red pixel on top stroke at (30,10)")
	}

	// Interior should be empty
	_, _, _, a2 := img.At(35, 35).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel in interior of stroked rect at (35,35)")
	}
}

func TestGGCanvasSaveRestore(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Save()
	c.Translate(50, 50)
	c.Restore()

	paint := &core.Paint{Color: color.RGBA{G: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, paint)
	img := c.Target()

	_, g, _, _ := img.At(5, 5).RGBA()
	if g == 0 {
		t.Error("expected green pixel at original position after restore")
	}
}

func TestGGCanvasTranslate(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Translate(20, 20)
	paint := &core.Paint{Color: color.RGBA{B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, paint)
	img := c.Target()

	// Rect should be at (20, 20) due to translate
	_, _, b, _ := img.At(25, 25).RGBA()
	if b == 0 {
		t.Error("expected blue pixel at translated position (25,25)")
	}

	// Original position should be empty
	_, _, b2, a2 := img.At(5, 5).RGBA()
	if b2 != 0 && a2 != 0 {
		t.Error("expected no blue pixel at original position (5,5)")
	}
}

func TestGGCanvasClipRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.ClipRect(core.Rect{X: 20, Y: 20, Width: 30, Height: 30})
	paint := &core.Paint{Color: color.RGBA{R: 255, G: 255, A: 255}, DrawStyle: core.PaintFill}
	// Draw a rect that extends beyond the clip
	c.DrawRect(core.Rect{X: 0, Y: 0, Width: 100, Height: 100}, paint)
	img := c.Target()

	// Inside clip: should have the color
	r, g, _, a := img.At(30, 30).RGBA()
	if r == 0 || g == 0 || a == 0 {
		t.Error("expected yellow pixel inside clip at (30,30)")
	}

	// Outside clip: should be transparent
	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel outside clip at (5,5)")
	}
}

func TestGGCanvasDrawCircle(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 128, G: 128, B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawCircle(50, 50, 20, paint)
	img := c.Target()

	// Center should be filled
	_, _, b, a := img.At(50, 50).RGBA()
	if b == 0 || a == 0 {
		t.Error("expected non-zero blue pixel at circle center (50,50)")
	}

	// Far outside the circle
	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel outside circle at (5,5)")
	}
}

func TestGGCanvasDrawLine(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintStroke}
	c.DrawLine(0, 0, 99, 99, paint)
	img := c.Target()

	// Diagonal line: pixel near the middle
	r, _, _, a := img.At(50, 50).RGBA()
	if r == 0 || a == 0 {
		t.Error("expected red pixel on the diagonal line at (50,50)")
	}
}

func TestGGCanvasDrawRoundRect(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	paint := &core.Paint{Color: color.RGBA{G: 200, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRoundRect(core.Rect{X: 10, Y: 10, Width: 60, Height: 60}, 10, paint)
	img := c.Target()

	// Center should be filled
	_, g, _, a := img.At(40, 40).RGBA()
	if g == 0 || a == 0 {
		t.Error("expected green pixel at center of round rect (40,40)")
	}

	// A corner pixel that is outside the rounded corner should be empty.
	// (10,10) is the exact corner — with radius 10, this pixel is outside the curve.
	_, _, _, a2 := img.At(10, 10).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel at rounded corner (10,10)")
	}
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
	img := c.Target()

	r, _, _, a := img.At(25, 25).RGBA()
	if r == 0 || a == 0 {
		t.Error("expected red pixel from drawn image at (25,25)")
	}

	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel outside drawn image at (5,5)")
	}
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
	img := c.Target()

	r, _, _, a := img.At(22, 22).RGBA()
	if r == 0 || a == 0 {
		t.Error("expected red pixel at (22,22) after double translate")
	}

	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel at (5,5) after double translate")
	}
}

func TestGGCanvasClipWithTranslate(t *testing.T) {
	c := NewGGCanvas(100, 100, nil)
	c.Translate(10, 10)
	c.ClipRect(core.Rect{X: 0, Y: 0, Width: 20, Height: 20})
	// Clip should be at (10, 10) to (30, 30) in image space
	paint := &core.Paint{Color: color.RGBA{B: 255, A: 255}, DrawStyle: core.PaintFill}
	c.DrawRect(core.Rect{X: -100, Y: -100, Width: 500, Height: 500}, paint)
	img := c.Target()

	// Inside clip
	_, _, b, a := img.At(20, 20).RGBA()
	if b == 0 || a == 0 {
		t.Error("expected blue pixel inside clip at (20,20)")
	}

	// Outside clip
	_, _, _, a2 := img.At(5, 5).RGBA()
	if a2 != 0 {
		t.Error("expected transparent pixel outside clip at (5,5)")
	}
}
