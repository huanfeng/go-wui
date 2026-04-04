package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
)

// Switch is a toggle switch widget with on/off state.
type Switch struct {
	BaseView
	on        bool
	onChanged func(on bool)
}

// NewSwitch creates a new Switch in the off state.
func NewSwitch() *Switch {
	sw := &Switch{}
	sw.node = initNode("Switch", sw)
	sw.node.SetPainter(&switchPainter{sw: sw})
	sw.node.SetHandler(&switchHandler{sw: sw})
	sw.node.SetStyle(&core.Style{})
	sw.node.SetData("checked", false)
	return sw
}

// IsOn reports whether the switch is currently on.
func (sw *Switch) IsOn() bool {
	return sw.on
}

// SetOn sets the on/off state.
func (sw *Switch) SetOn(on bool) {
	sw.on = on
	sw.node.SetData("checked", on)
	sw.node.MarkDirty()
}

// SetOnChanged sets the callback invoked when the switch state changes.
func (sw *Switch) SetOnChanged(fn func(on bool)) {
	sw.onChanged = fn
}

// switchPainter handles measurement and painting of the switch.
type switchPainter struct {
	sw *Switch
}

const (
	switchTrackWidth  = 44.0 // track width in dp
	switchTrackHeight = 24.0 // track height in dp
	switchThumbRadius = 10.0 // thumb radius in dp (20dp diameter)
	switchThumbInset  = 2.0  // inset from track edge
)

func (p *switchPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	scale := getDPIScale(node)
	w := switchTrackWidth * scale
	h := switchTrackHeight * scale

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

func (p *switchPainter) Paint(node *core.Node, canvas core.Canvas) {
	b := node.Bounds()
	scale := getDPIScale(node)
	thumbRadius := switchThumbRadius * scale
	thumbInset := switchThumbInset * scale

	// Track colors
	var trackColor color.RGBA
	if p.sw.on {
		trackColor = core.ParseColor("#1976D2") // primary color when on
	} else {
		trackColor = core.ParseColor("#BDBDBD") // gray when off
	}

	// Draw track (pill shape: rounded rect with radius = height/2)
	trackRect := core.Rect{Width: b.Width, Height: b.Height}
	trackRadius := b.Height / 2
	trackPaint := &core.Paint{Color: trackColor, DrawStyle: core.PaintFill}
	canvas.DrawRoundRect(trackRect, trackRadius, trackPaint)

	// Draw thumb (white circle)
	thumbColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	thumbPaint := &core.Paint{Color: thumbColor, DrawStyle: core.PaintFill}

	thumbCY := b.Height / 2
	var thumbCX float64
	if p.sw.on {
		// Thumb on right side
		thumbCX = b.Width - thumbInset - thumbRadius
	} else {
		// Thumb on left side
		thumbCX = thumbInset + thumbRadius
	}

	canvas.DrawCircle(thumbCX, thumbCY, thumbRadius, thumbPaint)
}

// switchHandler handles click events to toggle the switch state.
type switchHandler struct {
	core.DefaultHandler
	sw      *Switch
	pressed bool
}

func (h *switchHandler) OnEvent(node *core.Node, event core.Event) bool {
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
			newOn := !h.sw.on
			h.sw.SetOn(newOn)
			if h.sw.onChanged != nil {
				h.sw.onChanged(newOn)
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
