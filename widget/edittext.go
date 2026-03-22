package widget

import (
	"image/color"

	"gowui/core"
)

// EditText is a text-input widget backed by a platform-native EDIT control.
// The widget itself draws a placeholder border/background; the actual text
// editing is handled by the native control overlaid on top via AttachToNode.
type EditText struct {
	BaseView
	placeholder string
}

// NewEditText creates a new EditText with the given hint (placeholder) text.
func NewEditText(hint string) *EditText {
	et := &EditText{placeholder: hint}
	et.node = initNode("EditText", et)
	et.node.SetPainter(&editTextPainter{})
	et.node.SetStyle(&core.Style{
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	et.node.SetData("hint", hint)
	return et
}

// SetText stores text data on the node.
func (et *EditText) SetText(text string) {
	et.node.SetData("text", text)
	et.node.MarkDirty()
}

// GetText returns the current text stored on the node.
func (et *EditText) GetText() string {
	return et.node.GetDataString("text")
}

// SetHint updates the placeholder/hint text.
func (et *EditText) SetHint(hint string) {
	et.placeholder = hint
	et.node.SetData("hint", hint)
	et.node.MarkDirty()
}

// GetHint returns the current hint text.
func (et *EditText) GetHint() string {
	return et.placeholder
}

// ---------- Painter ----------

// editTextPainter draws the EditText background, border, and hint text
// when no native EDIT control is overlaid.
type editTextPainter struct{}

func (p *editTextPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	s := node.GetStyle()
	fontSize := 14.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}

	// Default size: fill available width, height ≈ font + padding.
	w := 200.0
	h := fontSize*1.4 + 16 // ~40dp at default font size

	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	} else if ws.Mode == core.MeasureModeAtMost && w > ws.Size {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	} else if hs.Mode == core.MeasureModeAtMost && h > hs.Size {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *editTextPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// Background (white).
	bgColor := s.BackgroundColor
	if bgColor.A == 0 {
		bgColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	bgPaint := &core.Paint{Color: bgColor, DrawStyle: core.PaintFill}
	if s.CornerRadius > 0 {
		canvas.DrawRoundRect(localRect, s.CornerRadius, bgPaint)
	} else {
		canvas.DrawRect(localRect, bgPaint)
	}

	// Border (gray).
	if s.BorderWidth > 0 {
		borderColor := s.BorderColor
		if borderColor.A == 0 {
			borderColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		}
		borderPaint := &core.Paint{
			Color:       borderColor,
			DrawStyle:   core.PaintStroke,
			StrokeWidth: s.BorderWidth,
		}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, borderPaint)
		} else {
			canvas.DrawRect(localRect, borderPaint)
		}
	}

	// Hint text (light gray) — only when there is no actual text.
	text := node.GetDataString("text")
	if text == "" {
		hint := node.GetDataString("hint")
		if hint != "" {
			fontSize := 14.0
			if s.FontSize > 0 {
				fontSize = s.FontSize
			}
			hintPaint := &core.Paint{
				Color:    color.RGBA{R: 180, G: 180, B: 180, A: 255},
				FontSize: fontSize,
			}
			// Left-aligned, vertically centered.
			textSize := canvas.MeasureText(hint, hintPaint)
			x := 8.0 // left padding
			y := (b.Height - textSize.Height) / 2
			canvas.DrawText(hint, x, y, hintPaint)
		}
	}
}
