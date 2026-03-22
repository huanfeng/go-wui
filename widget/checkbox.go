package widget

import (
	"image/color"

	"gowui/core"
)

// CheckBox is a checkable box with a text label.
type CheckBox struct {
	BaseView
	checked   bool
	onChanged func(checked bool)
}

// NewCheckBox creates a new CheckBox with the given label text.
func NewCheckBox(text string) *CheckBox {
	cb := &CheckBox{}
	cb.node = initNode("CheckBox", cb)
	cb.node.SetPainter(&checkBoxPainter{cb: cb})
	cb.node.SetHandler(&checkBoxHandler{cb: cb})
	cb.node.SetStyle(&core.Style{
		FontSize:  14,
		TextColor: color.RGBA{R: 33, G: 33, B: 33, A: 255},
	})
	cb.node.SetData("text", text)
	cb.node.SetData("checked", false)
	return cb
}

// IsChecked reports whether the checkbox is currently checked.
func (cb *CheckBox) IsChecked() bool {
	return cb.checked
}

// SetChecked sets the checked state.
func (cb *CheckBox) SetChecked(checked bool) {
	cb.checked = checked
	cb.node.SetData("checked", checked)
	cb.node.MarkDirty()
}

// SetOnCheckedChanged sets the callback invoked when the checked state changes.
func (cb *CheckBox) SetOnCheckedChanged(fn func(checked bool)) {
	cb.onChanged = fn
}

// SetText updates the label text.
func (cb *CheckBox) SetText(text string) {
	cb.node.SetData("text", text)
	cb.node.MarkDirty()
}

// GetText returns the current label text.
func (cb *CheckBox) GetText() string {
	return cb.node.GetDataString("text")
}

// checkBoxPainter handles measurement and painting of the checkbox.
type checkBoxPainter struct {
	cb *CheckBox
}

const (
	checkBoxSize = 20.0 // box size in dp
	checkBoxGap  = 8.0  // gap between box and text
)

func (p *checkBoxPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	text := node.GetDataString("text")
	s := node.GetStyle()
	fontSize := 14.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}

	charWidth := fontSize * 0.6
	textWidth := float64(len([]rune(text))) * charWidth

	w := checkBoxSize + checkBoxGap + textWidth
	h := max(checkBoxSize, fontSize*1.4)

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

func (p *checkBoxPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()

	// Position box vertically centered
	boxX := 0.0
	boxY := (b.Height - checkBoxSize) / 2

	primaryColor := core.ParseColor("#1976D2")

	if p.cb.checked {
		// Filled box with primary color
		fillPaint := &core.Paint{Color: primaryColor, DrawStyle: core.PaintFill}
		boxRect := core.Rect{X: boxX, Y: boxY, Width: checkBoxSize, Height: checkBoxSize}
		canvas.DrawRoundRect(boxRect, 3, fillPaint)

		// Draw checkmark (white)
		checkPaint := &core.Paint{
			Color:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			DrawStyle:   core.PaintStroke,
			StrokeWidth: 2,
		}
		// Short leg: bottom-left to bottom-center
		canvas.DrawLine(boxX+4, boxY+10, boxX+8, boxY+14, checkPaint)
		// Long leg: bottom-center to top-right
		canvas.DrawLine(boxX+8, boxY+14, boxX+16, boxY+5, checkPaint)
	} else {
		// Empty box outline
		borderPaint := &core.Paint{
			Color:       color.RGBA{R: 117, G: 117, B: 117, A: 255},
			DrawStyle:   core.PaintStroke,
			StrokeWidth: 2,
		}
		boxRect := core.Rect{X: boxX, Y: boxY, Width: checkBoxSize, Height: checkBoxSize}
		canvas.DrawRoundRect(boxRect, 3, borderPaint)
	}

	// Draw label text to the right of the box
	text := node.GetDataString("text")
	if text != "" {
		fontSize := 14.0
		if s.FontSize > 0 {
			fontSize = s.FontSize
		}
		textPaint := &core.Paint{
			Color:      s.TextColor,
			FontSize:   fontSize,
			FontFamily: s.FontFamily,
			FontWeight: s.FontWeight,
		}
		textX := checkBoxSize + checkBoxGap
		// Vertically center text
		textSize := canvas.MeasureText(text, textPaint)
		textY := (b.Height - textSize.Height) / 2
		canvas.DrawText(text, textX, textY, textPaint)
	}
}

// checkBoxHandler handles click events to toggle the checkbox state.
type checkBoxHandler struct {
	core.DefaultHandler
	cb      *CheckBox
	pressed bool
}

func (h *checkBoxHandler) OnEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		h.pressed = true
		node.MarkDirty()
		return true
	case core.ActionUp:
		if h.pressed && node.IsEnabled() {
			h.pressed = false
			newChecked := !h.cb.checked
			h.cb.SetChecked(newChecked)
			if h.cb.onChanged != nil {
				h.cb.onChanged(newChecked)
			}
		}
		return true
	case core.ActionCancel:
		h.pressed = false
		node.MarkDirty()
		return true
	}
	return false
}
