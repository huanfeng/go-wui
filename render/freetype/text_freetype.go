package freetype

import (
	"image"
	"image/color"
	"strings"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/huanfeng/go-wui/core"
)

// FreeTypeTextRenderer implements core.TextRenderer using Go's standard image
// and golang.org/x/image/font packages. Phase 1 uses basicfont.Face7x13 as
// the built-in bitmap font; proper TTF/OTF loading is a Phase 2 feature.
type FreeTypeTextRenderer struct {
	face     font.Face
	fontSize float64
}

// NewFreeTypeTextRenderer creates a new text renderer with basicfont.Face7x13.
func NewFreeTypeTextRenderer() *FreeTypeTextRenderer {
	return &FreeTypeTextRenderer{
		face:     basicfont.Face7x13,
		fontSize: 13,
	}
}

// SetFont records the requested font parameters. For Phase 1, the actual font
// face is always basicfont.Face7x13; only fontSize is used for scaling.
func (tr *FreeTypeTextRenderer) SetFont(fontFamily string, weight int, size float64) {
	if size > 0 {
		tr.fontSize = size
	}
	// Phase 1: always use basicfont. Font loading from files is a Phase 2 feature.
}

// MeasureText returns the bounding size of the given text string.
func (tr *FreeTypeTextRenderer) MeasureText(text string) core.Size {
	if text == "" {
		return core.Size{}
	}
	advance := font.MeasureString(tr.face, text)
	w := fixedToFloat(advance)
	scale := tr.fontSize / 13.0
	metrics := tr.face.Metrics()
	h := fixedToFloat(metrics.Ascent + metrics.Descent)
	return core.Size{Width: w * scale, Height: h * scale}
}

// DrawText renders text onto the canvas at (x, y) using the given paint.
// The y coordinate represents the baseline of the text.
func (tr *FreeTypeTextRenderer) DrawText(canvas core.Canvas, text string, x, y float64, paint *core.Paint) {
	if text == "" || canvas == nil {
		return
	}
	target := canvas.Target()
	if target == nil {
		return
	}

	clr := color.RGBA{A: 255} // default to opaque black
	if paint != nil && paint.Color.A != 0 {
		clr = paint.Color
	}

	// Draw text using font.Drawer.
	// Note: basicfont doesn't scale, so for Phase 1 we draw at native size
	// and accept the size mismatch. Proper font scaling is a Phase 2 feature.
	d := &font.Drawer{
		Dst:  target,
		Src:  image.NewUniform(clr),
		Face: tr.face,
		Dot:  fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(y)},
	}
	d.DrawString(text)
}

// CreateTextLayout performs greedy line-breaking and returns a TextLayoutResult
// describing how the text is split across multiple lines within maxWidth.
func (tr *FreeTypeTextRenderer) CreateTextLayout(text string, paint *core.Paint, maxWidth float64) *core.TextLayoutResult {
	if text == "" {
		return &core.TextLayoutResult{}
	}

	scale := tr.fontSize / 13.0
	var lines []core.TextLine
	var currentY float64

	metrics := tr.face.Metrics()
	lineHeight := fixedToFloat(metrics.Ascent+metrics.Descent) * scale
	ascent := fixedToFloat(metrics.Ascent) * scale

	// Simple greedy line-breaking.
	remaining := text
	for remaining != "" {
		line := remaining
		lineWidth := tr.measureRaw(line) * scale

		if maxWidth > 0 && lineWidth > maxWidth {
			breakIdx := tr.findBreakPoint(remaining, maxWidth, scale)
			if breakIdx <= 0 {
				// At least one character per line to guarantee progress.
				_, size := utf8.DecodeRuneInString(remaining)
				breakIdx = size
			}
			line = remaining[:breakIdx]
			lineWidth = tr.measureRaw(line) * scale
		}

		lines = append(lines, core.TextLine{
			Text:     strings.TrimRight(line, " \t"),
			Offset:   core.Point{X: 0, Y: currentY},
			Width:    lineWidth,
			Baseline: currentY + ascent,
		})
		currentY += lineHeight
		remaining = remaining[len(line):]
		remaining = strings.TrimLeft(remaining, " ") // skip leading spaces on new line
	}

	// Calculate total width.
	totalWidth := 0.0
	for _, l := range lines {
		if l.Width > totalWidth {
			totalWidth = l.Width
		}
	}

	return &core.TextLayoutResult{
		Lines:     lines,
		TotalSize: core.Size{Width: totalWidth, Height: currentY},
	}
}

// Close releases resources. basicfont doesn't need cleanup.
func (tr *FreeTypeTextRenderer) Close() {
	// basicfont.Face7x13 doesn't need cleanup.
}

// --- internal helpers ---

// measureRaw measures text width in raw (unscaled) pixels using the current face.
func (tr *FreeTypeTextRenderer) measureRaw(text string) float64 {
	return fixedToFloat(font.MeasureString(tr.face, text))
}

// findBreakPoint finds the byte index at which to break the text so that the
// rendered width (after scaling) does not exceed maxWidth. It prefers breaking
// at the last space/tab boundary.
func (tr *FreeTypeTextRenderer) findBreakPoint(text string, maxWidth float64, scale float64) int {
	lastSpace := -1
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		nextI := i + size
		w := tr.measureRaw(text[:nextI]) * scale
		if w > maxWidth {
			if lastSpace > 0 {
				return lastSpace + 1 // break after the last space
			}
			return i // break at current char (word too long)
		}
		if r == ' ' || r == '\t' {
			lastSpace = i
		}
		i = nextI
	}
	return len(text)
}

// fixedToFloat converts a fixed.Int26_6 value to float64.
func fixedToFloat(x fixed.Int26_6) float64 {
	return float64(x) / 64.0
}

// floatToFixed converts a float64 value to fixed.Int26_6.
func floatToFixed(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * 64.0)
}
