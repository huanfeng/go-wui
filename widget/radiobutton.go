package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/theme"
)

// RadioButton is a selectable circle with a text label.
// When part of a RadioGroup, only one RadioButton in the group can be selected at a time.
type RadioButton struct {
	BaseView
	selected  bool
	onChanged func(selected bool)
	group     *RadioGroup // back-reference to group (if any)
}

// NewRadioButton creates a new RadioButton with the given label text.
func NewRadioButton(text string) *RadioButton {
	rb := &RadioButton{}
	rb.node = initNode("RadioButton", rb)
	rb.node.SetPainter(&radioButtonPainter{rb: rb})
	rb.node.SetHandler(&radioButtonHandler{rb: rb})
	rb.node.SetStyle(&core.Style{
		FontSize:  14,
		TextColor: theme.CurrentColors().TextPrimary,
	})
	rb.node.SetData("text", text)
	rb.node.SetData("selected", false)
	return rb
}

// IsSelected reports whether the radio button is currently selected.
func (rb *RadioButton) IsSelected() bool {
	return rb.selected
}

// SetSelected sets the selected state.
// When called programmatically (not via group), it updates state and marks dirty.
// It does NOT notify the group to avoid infinite loops — the group manages selection.
func (rb *RadioButton) SetSelected(selected bool) {
	rb.selected = selected
	rb.node.SetData("selected", selected)
	rb.node.MarkDirty()
}

// SetOnSelectedChanged sets the callback invoked when the selected state changes.
func (rb *RadioButton) SetOnSelectedChanged(fn func(selected bool)) {
	rb.onChanged = fn
}

// SetText updates the label text.
func (rb *RadioButton) SetText(text string) {
	rb.node.SetData("text", text)
	rb.node.MarkDirty()
}

// GetText returns the current label text.
func (rb *RadioButton) GetText() string {
	return rb.node.GetDataString("text")
}

// radioButtonPainter handles measurement and painting of the radio button.
type radioButtonPainter struct {
	rb *RadioButton
}

const (
	radioCircleSize   = 20.0 // outer circle diameter in dp
	radioCircleRadius = 10.0 // outer circle radius
	radioInnerRadius  = 5.0  // inner filled circle radius
	radioGap          = 8.0  // gap between circle and text
)

func (p *radioButtonPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	scale := getDPIScale(node)
	circleSize := radioCircleSize * scale
	gap := radioGap * scale

	text := node.GetDataString("text")
	s := node.GetStyle()
	fontSize := 14.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}

	paint := &core.Paint{FontSize: fontSize}
	if s != nil {
		paint.FontFamily = s.FontFamily
		paint.FontWeight = s.FontWeight
	}
	textSize := core.NodeMeasureText(node, text, paint)

	w := circleSize + gap + textSize.Width
	h := max(circleSize, textSize.Height)

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

func (p *radioButtonPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	scale := getDPIScale(node)
	circleRadius := radioCircleRadius * scale
	innerRadius := radioInnerRadius * scale
	circleSize := radioCircleSize * scale
	gap := radioGap * scale

	// Center the circle vertically
	cx := circleRadius
	cy := b.Height / 2

	primaryColor := theme.CurrentColors().Primary

	// Draw outer circle (stroke only)
	outerPaint := &core.Paint{
		Color:       color.RGBA{R: 117, G: 117, B: 117, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 2 * scale,
	}
	if p.rb.selected {
		outerPaint.Color = primaryColor
	}
	canvas.DrawCircle(cx, cy, circleRadius, outerPaint)

	// If selected, draw inner filled circle
	if p.rb.selected {
		innerPaint := &core.Paint{
			Color:     primaryColor,
			DrawStyle: core.PaintFill,
		}
		canvas.DrawCircle(cx, cy, innerRadius, innerPaint)
	}

	// Draw label text to the right of the circle
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
		textX := circleSize + gap
		// Vertically center text
		textSize := canvas.MeasureText(text, textPaint)
		textY := (b.Height - textSize.Height) / 2
		canvas.DrawText(text, textX, textY, textPaint)
	}
}

// radioButtonHandler handles click events to select the radio button.
type radioButtonHandler struct {
	core.DefaultHandler
	rb      *RadioButton
	pressed bool
}

func (h *radioButtonHandler) OnEvent(node *core.Node, event core.Event) bool {
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
			if h.rb.group != nil {
				// Let the group manage selection (mutual exclusion)
				h.rb.group.selectButton(h.rb)
			} else {
				// Standalone radio button: toggle selected
				newSelected := !h.rb.selected
				h.rb.SetSelected(newSelected)
				if h.rb.onChanged != nil {
					h.rb.onChanged(newSelected)
				}
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
