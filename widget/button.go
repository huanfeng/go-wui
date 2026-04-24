package widget

import (
	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/theme"
)

// ButtonState represents the visual state of a button.
type ButtonState int

const (
	ButtonStateNormal  ButtonState = iota
	ButtonStateHovered
	ButtonStatePressed
)

// Button is a clickable widget that displays text with state-dependent styling.
type Button struct {
	BaseView
	state   ButtonState
	onClick func(core.View)
}

// NewButton creates a new Button with the given text and click handler.
func NewButton(text string, onClick func(core.View)) *Button {
	btn := &Button{onClick: onClick}
	btn.node = initNode("Button", btn)
	btn.node.SetPainter(&buttonPainter{btn: btn})
	btn.node.SetHandler(&buttonHandler{btn: btn})
	c := theme.CurrentColors()
	btn.node.SetStyle(&core.Style{
		FontSize:        16,
		BackgroundColor: c.Primary,
		TextColor:       c.TextOnPrimary,
		CornerRadius:    4,
	})
	btn.node.SetData("text", text)
	return btn
}

// SetText updates the button label.
func (btn *Button) SetText(text string) {
	btn.node.SetData("text", text)
	btn.node.MarkDirty()
}

// GetText returns the current button label.
func (btn *Button) GetText() string {
	return btn.node.GetDataString("text")
}

// SetOnClickListener sets the callback invoked when the button is clicked.
func (btn *Button) SetOnClickListener(fn func(core.View)) {
	btn.onClick = fn
}

// State returns the current visual state of the button.
func (btn *Button) State() ButtonState {
	return btn.state
}

// buttonPainter draws the button with state-dependent colors.
type buttonPainter struct {
	btn *Button
}

func (p *buttonPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	text := node.GetDataString("text")
	s := node.GetStyle()
	fontSize := 16.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}

	paint := &core.Paint{FontSize: fontSize}
	if s != nil {
		paint.FontFamily = s.FontFamily
		paint.FontWeight = s.FontWeight
	}
	textSize := core.NodeMeasureText(node, text, paint)
	w := textSize.Width + 32 // horizontal padding
	h := textSize.Height + 16 // vertical padding

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

func (p *buttonPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// State-dependent background
	bgColor := s.BackgroundColor
	switch p.btn.state {
	case ButtonStateHovered:
		// Slightly lighter
		bgColor.R = uint8(min(int(bgColor.R)+20, 255))
		bgColor.G = uint8(min(int(bgColor.G)+20, 255))
		bgColor.B = uint8(min(int(bgColor.B)+20, 255))
	case ButtonStatePressed:
		// Slightly darker
		bgColor.R = uint8(max(int(bgColor.R)-30, 0))
		bgColor.G = uint8(max(int(bgColor.G)-30, 0))
		bgColor.B = uint8(max(int(bgColor.B)-30, 0))
	}

	bgPaint := &core.Paint{Color: bgColor, DrawStyle: core.PaintFill}
	if s.CornerRadius > 0 {
		canvas.DrawRoundRect(localRect, s.CornerRadius, bgPaint)
	} else {
		canvas.DrawRect(localRect, bgPaint)
	}

	// Draw border if set
	if s.BorderWidth > 0 && s.BorderColor.A > 0 {
		borderPaint := &core.Paint{Color: s.BorderColor, DrawStyle: core.PaintStroke, StrokeWidth: s.BorderWidth}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, borderPaint)
		} else {
			canvas.DrawRect(localRect, borderPaint)
		}
	}

	// Draw text centered
	text := node.GetDataString("text")
	if text != "" {
		fontSize := 16.0
		if s.FontSize > 0 {
			fontSize = s.FontSize
		}
		textPaint := &core.Paint{
			Color:      s.TextColor,
			FontSize:   fontSize,
			FontFamily: s.FontFamily,
			FontWeight: s.FontWeight,
		}
		// Center text using actual measurement
		textSize := canvas.MeasureText(text, textPaint)
		x := (b.Width - textSize.Width) / 2
		y := (b.Height - textSize.Height) / 2
		canvas.DrawText(text, x, y, textPaint)
	}
}

// buttonHandler handles pointer events for click/hover/press state changes.
type buttonHandler struct {
	core.DefaultHandler
	btn *Button
}

func (h *buttonHandler) OnEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		h.btn.state = ButtonStatePressed
		node.MarkDirty()
		return true
	case core.ActionUp:
		wasPressed := h.btn.state == ButtonStatePressed
		h.btn.state = ButtonStateNormal
		node.MarkDirty()
		if wasPressed && h.btn.onClick != nil && node.IsEnabled() {
			h.btn.onClick(h.btn)
		}
		return true
	case core.ActionHoverEnter:
		h.btn.state = ButtonStateHovered
		node.MarkDirty()
		return true
	case core.ActionHoverExit, core.ActionCancel:
		h.btn.state = ButtonStateNormal
		node.MarkDirty()
		return true
	}
	return false
}
