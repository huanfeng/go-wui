package widget

import "gowui/core"

// TextView displays a single block of text.
type TextView struct {
	BaseView
}

// NewTextView creates a new TextView with the given initial text.
func NewTextView(text string) *TextView {
	tv := &TextView{}
	tv.node = initNode("TextView", tv)
	tv.node.SetPainter(&textViewPainter{})
	tv.node.SetStyle(&core.Style{FontSize: 14})
	tv.node.SetData("text", text)
	return tv
}

// SetText updates the displayed text.
func (tv *TextView) SetText(text string) {
	tv.node.SetData("text", text)
	tv.node.MarkDirty()
}

// GetText returns the current text.
func (tv *TextView) GetText() string {
	return tv.node.GetDataString("text")
}

// SetTextSize sets the font size for the text.
func (tv *TextView) SetTextSize(size float64) {
	if s := tv.node.GetStyle(); s != nil {
		s.FontSize = size
		tv.node.MarkDirty()
	}
}

// textViewPainter measures and draws text.
type textViewPainter struct{}

func (p *textViewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	text := node.GetDataString("text")
	if text == "" {
		return core.Size{}
	}
	s := node.GetStyle()
	fontSize := 14.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}

	// Estimate text size: ~0.6 * fontSize per character width, fontSize * 1.4 for height.
	// This is a rough estimate — actual rendering uses TextRenderer.
	charWidth := fontSize * 0.6
	w := float64(len([]rune(text))) * charWidth
	h := fontSize * 1.4

	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if ws.Mode == core.MeasureModeAtMost && w > ws.Size {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	if hs.Mode == core.MeasureModeAtMost && h > hs.Size {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *textViewPainter) Paint(node *core.Node, canvas core.Canvas) {
	text := node.GetDataString("text")
	if text == "" {
		return
	}
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()

	// Draw background first (if set)
	if s.BackgroundColor.A > 0 {
		bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, bgPaint)
	}

	// Draw text with gravity support
	fontSize := 14.0
	if s.FontSize > 0 {
		fontSize = s.FontSize
	}
	paint := &core.Paint{
		Color:      s.TextColor,
		FontSize:   fontSize,
		FontFamily: s.FontFamily,
		FontWeight: s.FontWeight,
	}

	// Measure actual text size for gravity positioning
	textSize := canvas.MeasureText(text, paint)

	// Horizontal position based on gravity
	x := 0.0
	switch s.Gravity {
	case core.GravityCenter:
		x = (b.Width - textSize.Width) / 2
	case core.GravityEnd:
		x = b.Width - textSize.Width
	}

	// Vertical centering
	y := (b.Height - textSize.Height) / 2

	canvas.DrawText(text, x, y, paint)
}
