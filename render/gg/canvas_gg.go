package gg

import (
	"image"
	"image/color"
	"image/draw"

	foggg "github.com/fogleman/gg"

	"gowui/core"
)

// clipRect tracks the current clip rectangle in absolute (image) coordinates.
type clipRect struct {
	x1, y1, x2, y2 float64 // min/max corners
	active          bool
}

// GGCanvas implements core.Canvas using the fogleman/gg library
// for anti-aliased 2D drawing.
type GGCanvas struct {
	dc           *foggg.Context
	textRenderer core.TextRenderer
	width        int
	height       int

	// Manual translate tracking for operations that bypass gg's transform
	// (DrawText and DrawImage write directly to the raw *image.RGBA).
	txStack []point2
	tx, ty  float64 // accumulated translate offset

	// Manual clip tracking for DrawText/DrawImage (they bypass gg's clip).
	clip      clipRect
	clipStack []clipRect
}

type point2 struct{ x, y float64 }

// NewGGCanvas creates a new GGCanvas with the given dimensions.
// textRenderer may be nil; text operations become no-ops in that case.
func NewGGCanvas(width, height int, textRenderer core.TextRenderer) *GGCanvas {
	dc := foggg.NewContext(width, height)
	return &GGCanvas{
		dc:           dc,
		textRenderer: textRenderer,
		width:        width,
		height:       height,
		clip:         clipRect{x1: 0, y1: 0, x2: float64(width), y2: float64(height), active: false},
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

	// Apply accumulated translate offset since we write directly to raw pixels.
	dx := dst.X + c.tx
	dy := dst.Y + c.ty
	target := c.targetRGBA()

	// Clip the draw region to the current clip rect
	drawRect := image.Rect(int(dx), int(dy), int(dx+dst.Width), int(dy+dst.Height))
	if c.clip.active {
		cr := image.Rect(int(c.clip.x1), int(c.clip.y1), int(c.clip.x2), int(c.clip.y2))
		drawRect = drawRect.Intersect(cr)
	}

	srcPoint := image.Point{X: drawRect.Min.X - int(dx), Y: drawRect.Min.Y - int(dy)}
	draw.Draw(target, drawRect, tmp, srcPoint, draw.Over)
}

// DrawText delegates to the injected TextRenderer. No-op if nil.
// Applies accumulated translate offset since TextRenderer writes directly
// to the raw *image.RGBA, bypassing gg's internal transform.
func (c *GGCanvas) DrawText(text string, x, y float64, paint *core.Paint) {
	if c.textRenderer == nil || paint == nil {
		return
	}

	absX := x + c.tx
	absY := y + c.ty

	// Skip text entirely if it's outside the clip rect
	if c.clip.active {
		fontSize := paint.FontSize
		if fontSize <= 0 {
			fontSize = 14
		}
		// Estimate text height for clip check
		textH := fontSize * 1.5
		if absY+textH < c.clip.y1 || absY > c.clip.y2 {
			return // completely above or below clip
		}
		if absX > c.clip.x2 {
			return // completely to the right of clip
		}
	}

	c.textRenderer.SetFont(paint.FontFamily, paint.FontWeight, paint.FontSize)

	if c.clip.active {
		// Draw text to a temporary image, then copy only the clipped region
		c.drawTextClipped(text, absX, absY, paint)
	} else {
		c.textRenderer.DrawText(c, text, absX, absY, paint)
	}
}

// drawTextClipped renders text into a temporary buffer and copies only the
// portion that falls within the active clip rect to the canvas target.
func (c *GGCanvas) drawTextClipped(text string, absX, absY float64, paint *core.Paint) {
	target := c.targetRGBA()
	bounds := target.Bounds()

	// Save the region that will be drawn to, render text, then mask
	// pixels outside the clip rect by restoring the saved pixels.
	cr := image.Rect(int(c.clip.x1), int(c.clip.y1), int(c.clip.x2), int(c.clip.y2))
	cr = cr.Intersect(bounds)
	if cr.Empty() {
		return
	}

	// Estimate the text bounding box
	textSize := c.textRenderer.MeasureText(text)
	textRect := image.Rect(int(absX), int(absY), int(absX+textSize.Width+2), int(absY+textSize.Height+2))
	textRect = textRect.Intersect(bounds)
	if textRect.Empty() {
		return
	}

	// Find pixels that are inside textRect but OUTSIDE clip — save them
	type savedPixel struct {
		x, y int
		c    color.RGBA
	}
	var saved []savedPixel
	for py := textRect.Min.Y; py < textRect.Max.Y; py++ {
		for px := textRect.Min.X; px < textRect.Max.X; px++ {
			if !image.Pt(px, py).In(cr) {
				idx := target.PixOffset(px, py)
				if idx+3 < len(target.Pix) {
					saved = append(saved, savedPixel{
						x: px, y: py,
						c: color.RGBA{
							R: target.Pix[idx],
							G: target.Pix[idx+1],
							B: target.Pix[idx+2],
							A: target.Pix[idx+3],
						},
					})
				}
			}
		}
	}

	// Draw text normally
	c.textRenderer.DrawText(c, text, absX, absY, paint)

	// Restore pixels outside clip rect
	for _, sp := range saved {
		idx := target.PixOffset(sp.x, sp.y)
		if idx+3 < len(target.Pix) {
			target.Pix[idx] = sp.c.R
			target.Pix[idx+1] = sp.c.G
			target.Pix[idx+2] = sp.c.B
			target.Pix[idx+3] = sp.c.A
		}
	}
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
	c.txStack = append(c.txStack, point2{c.tx, c.ty})
	c.clipStack = append(c.clipStack, c.clip)
}

// Restore pops the most recent drawing state from the stack.
func (c *GGCanvas) Restore() {
	c.dc.Pop()
	if len(c.txStack) > 0 {
		top := c.txStack[len(c.txStack)-1]
		c.txStack = c.txStack[:len(c.txStack)-1]
		c.tx, c.ty = top.x, top.y
	}
	if len(c.clipStack) > 0 {
		c.clip = c.clipStack[len(c.clipStack)-1]
		c.clipStack = c.clipStack[:len(c.clipStack)-1]
	}
}

// Translate accumulates a translation offset.
func (c *GGCanvas) Translate(dx, dy float64) {
	c.dc.Translate(dx, dy)
	c.tx += dx
	c.ty += dy
}

// ClipRect intersects the current clip with the given rectangle.
func (c *GGCanvas) ClipRect(rect core.Rect) {
	c.dc.DrawRectangle(rect.X, rect.Y, rect.Width, rect.Height)
	c.dc.Clip()

	// Track clip rect in absolute (image) coordinates for DrawText/DrawImage
	absX1 := rect.X + c.tx
	absY1 := rect.Y + c.ty
	absX2 := absX1 + rect.Width
	absY2 := absY1 + rect.Height

	if c.clip.active {
		// Intersect with existing clip
		if absX1 < c.clip.x1 {
			absX1 = c.clip.x1
		}
		if absY1 < c.clip.y1 {
			absY1 = c.clip.y1
		}
		if absX2 > c.clip.x2 {
			absX2 = c.clip.x2
		}
		if absY2 > c.clip.y2 {
			absY2 = c.clip.y2
		}
	}

	c.clip = clipRect{x1: absX1, y1: absY1, x2: absX2, y2: absY2, active: true}
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
