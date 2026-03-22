package gg

import (
	"image"
	"image/color"
	"image/draw"

	foggg "github.com/fogleman/gg"

	"gowui/core"
)

// GGCanvas implements core.Canvas using the fogleman/gg library
// for anti-aliased 2D drawing.
type GGCanvas struct {
	dc           *foggg.Context
	textRenderer core.TextRenderer
	width        int
	height       int
}

// NewGGCanvas creates a new GGCanvas with the given dimensions.
// textRenderer may be nil; text operations become no-ops in that case.
func NewGGCanvas(width, height int, textRenderer core.TextRenderer) *GGCanvas {
	dc := foggg.NewContext(width, height)
	return &GGCanvas{
		dc:           dc,
		textRenderer: textRenderer,
		width:        width,
		height:       height,
	}
}

// ---------- Drawing primitives ----------

// DrawRect draws a filled or stroked rectangle.
func (c *GGCanvas) DrawRect(rect core.Rect, paint *core.Paint) {
	if paint == nil {
		return
	}
	c.dc.DrawRectangle(rect.X, rect.Y, rect.Width, rect.Height)
	c.applyPaint(paint)
}

// DrawRoundRect draws a rounded rectangle.
func (c *GGCanvas) DrawRoundRect(rect core.Rect, radius float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	if radius <= 0 {
		c.DrawRect(rect, paint)
		return
	}
	c.dc.DrawRoundedRectangle(rect.X, rect.Y, rect.Width, rect.Height, radius)
	c.applyPaint(paint)
}

// DrawCircle draws a filled or stroked circle.
func (c *GGCanvas) DrawCircle(cx, cy, radius float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	c.dc.DrawCircle(cx, cy, radius)
	c.applyPaint(paint)
}

// DrawLine draws a line between two points.
func (c *GGCanvas) DrawLine(x1, y1, x2, y2 float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	c.dc.DrawLine(x1, y1, x2, y2)
	c.dc.SetColor(paint.Color)
	sw := paint.StrokeWidth
	if sw <= 0 {
		sw = 1
	}
	c.dc.SetLineWidth(sw)
	c.dc.Stroke()
}

// DrawImage draws an ImageResource into the destination rectangle.
func (c *GGCanvas) DrawImage(img *core.ImageResource, dst core.Rect) {
	if img == nil || img.Image == nil {
		return
	}
	srcBounds := img.Image.Bounds()

	// Simple nearest-neighbour scaling into a temp RGBA, then draw.
	tmp := image.NewRGBA(image.Rect(0, 0, int(dst.Width), int(dst.Height)))
	for py := 0; py < int(dst.Height); py++ {
		for px := 0; px < int(dst.Width); px++ {
			sx := srcBounds.Min.X + px*srcBounds.Dx()/int(dst.Width)
			sy := srcBounds.Min.Y + py*srcBounds.Dy()/int(dst.Height)
			if sx >= srcBounds.Max.X {
				sx = srcBounds.Max.X - 1
			}
			if sy >= srcBounds.Max.Y {
				sy = srcBounds.Max.Y - 1
			}
			tmp.Set(px, py, img.Image.At(sx, sy))
		}
	}

	// Draw scaled image onto the gg context's underlying image.
	target := c.targetRGBA()
	draw.Draw(target,
		image.Rect(int(dst.X), int(dst.Y), int(dst.X+dst.Width), int(dst.Y+dst.Height)),
		tmp, image.Point{}, draw.Over)
}

// DrawText delegates to the injected TextRenderer. No-op if nil.
func (c *GGCanvas) DrawText(text string, x, y float64, paint *core.Paint) {
	if c.textRenderer == nil || paint == nil {
		return
	}
	c.textRenderer.SetFont(paint.FontFamily, paint.FontWeight, paint.FontSize)
	c.textRenderer.DrawText(c, text, x, y, paint)
}

// MeasureText delegates to the injected TextRenderer. Returns zero Size if nil.
func (c *GGCanvas) MeasureText(text string, paint *core.Paint) core.Size {
	if c.textRenderer == nil || paint == nil {
		return core.Size{}
	}
	c.textRenderer.SetFont(paint.FontFamily, paint.FontWeight, paint.FontSize)
	return c.textRenderer.MeasureText(text)
}

// ---------- State management ----------

// Save pushes the current drawing state (transform, clip) onto the stack.
func (c *GGCanvas) Save() {
	c.dc.Push()
}

// Restore pops the most recent drawing state from the stack.
func (c *GGCanvas) Restore() {
	c.dc.Pop()
}

// Translate accumulates a translation offset.
func (c *GGCanvas) Translate(dx, dy float64) {
	c.dc.Translate(dx, dy)
}

// ClipRect intersects the current clip with the given rectangle.
func (c *GGCanvas) ClipRect(rect core.Rect) {
	c.dc.DrawRectangle(rect.X, rect.Y, rect.Width, rect.Height)
	c.dc.Clip()
}

// Target returns the underlying RGBA image.
func (c *GGCanvas) Target() *image.RGBA {
	return c.targetRGBA()
}

// ---------- Internal helpers ----------

// targetRGBA returns the gg context's image as *image.RGBA.
func (c *GGCanvas) targetRGBA() *image.RGBA {
	img := c.dc.Image()
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	// Fallback: copy into a new RGBA image.
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

// applyPaint applies fill/stroke based on the paint style.
func (c *GGCanvas) applyPaint(paint *core.Paint) {
	c.dc.SetColor(color.RGBA(paint.Color))
	switch paint.DrawStyle {
	case core.PaintFill:
		c.dc.Fill()
	case core.PaintStroke:
		sw := paint.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		c.dc.SetLineWidth(sw)
		c.dc.Stroke()
	case core.PaintFillAndStroke:
		c.dc.FillPreserve()
		sw := paint.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		c.dc.SetLineWidth(sw)
		c.dc.Stroke()
	}
}
