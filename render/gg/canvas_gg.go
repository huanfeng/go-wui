package gg

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"gowui/core"
)

// canvasState holds the transform and clip state for Save/Restore.
type canvasState struct {
	translateX, translateY float64
	clipRect               *core.Rect
}

// GGCanvas implements core.Canvas using Go's standard image/draw package
// with the fogleman/gg library available for future advanced rendering.
type GGCanvas struct {
	width, height int
	img           *image.RGBA
	textRenderer  core.TextRenderer
	stateStack    []canvasState
	currentState  canvasState
}

// NewGGCanvas creates a new GGCanvas with the given dimensions.
// textRenderer may be nil; text operations become no-ops in that case.
func NewGGCanvas(width, height int, textRenderer core.TextRenderer) *GGCanvas {
	return &GGCanvas{
		width:        width,
		height:       height,
		img:          image.NewRGBA(image.Rect(0, 0, width, height)),
		textRenderer: textRenderer,
	}
}

// ---------- Drawing primitives ----------

// DrawRect draws a filled or stroked rectangle.
func (c *GGCanvas) DrawRect(rect core.Rect, paint *core.Paint) {
	if paint == nil {
		return
	}
	r := c.applyTranslate(rect)

	switch paint.DrawStyle {
	case core.PaintFill, core.PaintFillAndStroke:
		c.fillRect(r, paint.Color)
	}
	if paint.DrawStyle == core.PaintStroke || paint.DrawStyle == core.PaintFillAndStroke {
		c.strokeRect(r, paint.Color, paint.StrokeWidth)
	}
}

// DrawRoundRect draws a rounded rectangle. Phase 1 simplifies corners
// to quarter-circle arcs drawn pixel by pixel.
func (c *GGCanvas) DrawRoundRect(rect core.Rect, radius float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	r := c.applyTranslate(rect)

	if radius <= 0 {
		c.DrawRect(rect, paint)
		return
	}

	// Clamp radius so it doesn't exceed half the smaller dimension.
	maxR := math.Min(r.Width/2, r.Height/2)
	if radius > maxR {
		radius = maxR
	}

	if paint.DrawStyle == core.PaintFill || paint.DrawStyle == core.PaintFillAndStroke {
		c.fillRoundRect(r, radius, paint.Color)
	}
	if paint.DrawStyle == core.PaintStroke || paint.DrawStyle == core.PaintFillAndStroke {
		// For Phase 1, stroke falls back to the filled outline approach.
		c.strokeRoundRect(r, radius, paint.Color, paint.StrokeWidth)
	}
}

// DrawCircle draws a filled or stroked circle.
func (c *GGCanvas) DrawCircle(cx, cy, radius float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	cx += c.currentState.translateX
	cy += c.currentState.translateY

	if paint.DrawStyle == core.PaintFill || paint.DrawStyle == core.PaintFillAndStroke {
		c.fillCircle(cx, cy, radius, paint.Color)
	}
	if paint.DrawStyle == core.PaintStroke || paint.DrawStyle == core.PaintFillAndStroke {
		c.strokeCircle(cx, cy, radius, paint.Color, paint.StrokeWidth)
	}
}

// DrawLine draws a line between two points using Bresenham's algorithm.
func (c *GGCanvas) DrawLine(x1, y1, x2, y2 float64, paint *core.Paint) {
	if paint == nil {
		return
	}
	x1 += c.currentState.translateX
	y1 += c.currentState.translateY
	x2 += c.currentState.translateX
	y2 += c.currentState.translateY

	c.bresenham(int(math.Round(x1)), int(math.Round(y1)),
		int(math.Round(x2)), int(math.Round(y2)), paint.Color)
}

// DrawImage draws an ImageResource into the destination rectangle.
func (c *GGCanvas) DrawImage(img *core.ImageResource, dst core.Rect) {
	if img == nil || img.Image == nil {
		return
	}
	d := c.applyTranslate(dst)
	srcBounds := img.Image.Bounds()

	// Simple nearest-neighbour scaling.
	for py := 0; py < int(d.Height); py++ {
		for px := 0; px < int(d.Width); px++ {
			dx := int(d.X) + px
			dy := int(d.Y) + py
			if !c.inClip(dx, dy) {
				continue
			}
			sx := srcBounds.Min.X + px*srcBounds.Dx()/int(d.Width)
			sy := srcBounds.Min.Y + py*srcBounds.Dy()/int(d.Height)
			if sx >= srcBounds.Max.X {
				sx = srcBounds.Max.X - 1
			}
			if sy >= srcBounds.Max.Y {
				sy = srcBounds.Max.Y - 1
			}
			c.img.Set(dx, dy, img.Image.At(sx, sy))
		}
	}
}

// DrawText delegates to the injected TextRenderer. No-op if nil.
func (c *GGCanvas) DrawText(text string, x, y float64, paint *core.Paint) {
	if c.textRenderer == nil || paint == nil {
		return
	}
	x += c.currentState.translateX
	y += c.currentState.translateY
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

// Save pushes the current transform/clip state onto the stack.
func (c *GGCanvas) Save() {
	clone := c.currentState
	if c.currentState.clipRect != nil {
		cr := *c.currentState.clipRect
		clone.clipRect = &cr
	}
	c.stateStack = append(c.stateStack, clone)
}

// Restore pops the most recent state from the stack.
func (c *GGCanvas) Restore() {
	n := len(c.stateStack)
	if n == 0 {
		return
	}
	c.currentState = c.stateStack[n-1]
	c.stateStack = c.stateStack[:n-1]
}

// Translate accumulates a translation offset.
func (c *GGCanvas) Translate(dx, dy float64) {
	c.currentState.translateX += dx
	c.currentState.translateY += dy
}

// ClipRect intersects the current clip with the given rectangle.
func (c *GGCanvas) ClipRect(rect core.Rect) {
	r := c.applyTranslate(rect)
	if c.currentState.clipRect != nil {
		intersected := c.currentState.clipRect.Intersect(r)
		c.currentState.clipRect = &intersected
	} else {
		c.currentState.clipRect = &r
	}
}

// Target returns the underlying RGBA image.
func (c *GGCanvas) Target() *image.RGBA {
	return c.img
}

// ---------- Internal helpers ----------

// applyTranslate returns a new Rect shifted by the current translation.
func (c *GGCanvas) applyTranslate(r core.Rect) core.Rect {
	return core.Rect{
		X:      r.X + c.currentState.translateX,
		Y:      r.Y + c.currentState.translateY,
		Width:  r.Width,
		Height: r.Height,
	}
}

// inClip reports whether pixel (px, py) falls within the current clip and image bounds.
func (c *GGCanvas) inClip(px, py int) bool {
	if px < 0 || py < 0 || px >= c.width || py >= c.height {
		return false
	}
	if cr := c.currentState.clipRect; cr != nil {
		return float64(px) >= cr.X && float64(px) < cr.X+cr.Width &&
			float64(py) >= cr.Y && float64(py) < cr.Y+cr.Height
	}
	return true
}

// setPixel sets a single pixel if it is within clip/bounds.
func (c *GGCanvas) setPixel(x, y int, clr color.RGBA) {
	if c.inClip(x, y) {
		c.img.SetRGBA(x, y, clr)
	}
}

// fillRect fills a rectangle with solid color.
func (c *GGCanvas) fillRect(r core.Rect, clr color.RGBA) {
	x0 := int(math.Round(r.X))
	y0 := int(math.Round(r.Y))
	x1 := int(math.Round(r.X + r.Width))
	y1 := int(math.Round(r.Y + r.Height))

	// Fast path: if no clip, use draw.Draw for the overlapping region.
	if c.currentState.clipRect == nil {
		// Clamp to image bounds.
		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 > c.width {
			x1 = c.width
		}
		if y1 > c.height {
			y1 = c.height
		}
		draw.Draw(c.img, image.Rect(x0, y0, x1, y1),
			image.NewUniform(clr), image.Point{}, draw.Src)
		return
	}

	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			c.setPixel(px, py, clr)
		}
	}
}

// strokeRect draws only the outline of a rectangle.
func (c *GGCanvas) strokeRect(r core.Rect, clr color.RGBA, strokeWidth float64) {
	if strokeWidth <= 0 {
		strokeWidth = 1
	}
	sw := int(math.Ceil(strokeWidth))

	x0 := int(math.Round(r.X))
	y0 := int(math.Round(r.Y))
	x1 := int(math.Round(r.X + r.Width))
	y1 := int(math.Round(r.Y + r.Height))

	// Top edge
	for s := 0; s < sw; s++ {
		for px := x0; px < x1; px++ {
			c.setPixel(px, y0+s, clr)
		}
	}
	// Bottom edge
	for s := 0; s < sw; s++ {
		for px := x0; px < x1; px++ {
			c.setPixel(px, y1-1-s, clr)
		}
	}
	// Left edge
	for s := 0; s < sw; s++ {
		for py := y0; py < y1; py++ {
			c.setPixel(x0+s, py, clr)
		}
	}
	// Right edge
	for s := 0; s < sw; s++ {
		for py := y0; py < y1; py++ {
			c.setPixel(x1-1-s, py, clr)
		}
	}
}

// fillRoundRect fills a rounded rectangle scanline by scanline.
func (c *GGCanvas) fillRoundRect(r core.Rect, radius float64, clr color.RGBA) {
	x0 := int(math.Round(r.X))
	y0 := int(math.Round(r.Y))
	x1 := int(math.Round(r.X + r.Width))
	y1 := int(math.Round(r.Y + r.Height))
	rad := int(math.Round(radius))

	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if c.inRoundRect(px, py, x0, y0, x1, y1, rad) {
				c.setPixel(px, py, clr)
			}
		}
	}
}

// strokeRoundRect draws the outline of a rounded rectangle.
func (c *GGCanvas) strokeRoundRect(r core.Rect, radius float64, clr color.RGBA, strokeWidth float64) {
	if strokeWidth <= 0 {
		strokeWidth = 1
	}
	sw := math.Ceil(strokeWidth)

	x0 := int(math.Round(r.X))
	y0 := int(math.Round(r.Y))
	x1 := int(math.Round(r.X + r.Width))
	y1 := int(math.Round(r.Y + r.Height))
	rad := int(math.Round(radius))

	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if !c.inRoundRect(px, py, x0, y0, x1, y1, rad) {
				continue
			}
			// Check if inside the inner rectangle (not on the border).
			innerRad := int(math.Round(radius - sw))
			if innerRad < 0 {
				innerRad = 0
			}
			ix0 := x0 + int(sw)
			iy0 := y0 + int(sw)
			ix1 := x1 - int(sw)
			iy1 := y1 - int(sw)
			if ix0 < ix1 && iy0 < iy1 && c.inRoundRect(px, py, ix0, iy0, ix1, iy1, innerRad) {
				continue
			}
			c.setPixel(px, py, clr)
		}
	}
}

// inRoundRect reports whether (px, py) falls within a rounded rectangle.
func (c *GGCanvas) inRoundRect(px, py, x0, y0, x1, y1, rad int) bool {
	if px < x0 || px >= x1 || py < y0 || py >= y1 {
		return false
	}

	// Check corners.
	corners := [4][2]int{
		{x0 + rad, y0 + rad}, // top-left
		{x1 - rad, y0 + rad}, // top-right (note: x1-rad not x1-rad-1 for center)
		{x0 + rad, y1 - rad}, // bottom-left
		{x1 - rad, y1 - rad}, // bottom-right
	}

	for i, center := range corners {
		var inCorner bool
		switch i {
		case 0:
			inCorner = px < x0+rad && py < y0+rad
		case 1:
			inCorner = px >= x1-rad && py < y0+rad
		case 2:
			inCorner = px < x0+rad && py >= y1-rad
		case 3:
			inCorner = px >= x1-rad && py >= y1-rad
		}
		if inCorner {
			dx := float64(px - center[0])
			dy := float64(py - center[1])
			if dx*dx+dy*dy > float64(rad*rad) {
				return false
			}
		}
	}

	return true
}

// fillCircle fills a circle using the midpoint test.
func (c *GGCanvas) fillCircle(cx, cy, radius float64, clr color.RGBA) {
	r := int(math.Ceil(radius))
	icx := int(math.Round(cx))
	icy := int(math.Round(cy))
	r2 := radius * radius

	for py := icy - r; py <= icy+r; py++ {
		for px := icx - r; px <= icx+r; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			if dx*dx+dy*dy <= r2 {
				c.setPixel(px, py, clr)
			}
		}
	}
}

// strokeCircle draws the outline of a circle.
func (c *GGCanvas) strokeCircle(cx, cy, radius float64, clr color.RGBA, strokeWidth float64) {
	if strokeWidth <= 0 {
		strokeWidth = 1
	}
	outerR := radius
	innerR := radius - strokeWidth
	if innerR < 0 {
		innerR = 0
	}
	outerR2 := outerR * outerR
	innerR2 := innerR * innerR

	r := int(math.Ceil(outerR))
	icx := int(math.Round(cx))
	icy := int(math.Round(cy))

	for py := icy - r; py <= icy+r; py++ {
		for px := icx - r; px <= icx+r; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			d2 := dx*dx + dy*dy
			if d2 <= outerR2 && d2 >= innerR2 {
				c.setPixel(px, py, clr)
			}
		}
	}
}

// bresenham draws a line using Bresenham's algorithm.
func (c *GGCanvas) bresenham(x0, y0, x1, y1 int, clr color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy

	for {
		c.setPixel(x0, y0, clr)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
