package widget

import (
	"image/color"

	"github.com/huanfeng/go-wui/core"
)

const (
	seekBarTrackHeight = 4.0
	seekBarThumbRadius = 8.0
)

// SeekBar is a horizontal slider that allows selecting a value between 0.0 and 1.0.
// Modeled after Android's SeekBar.
type SeekBar struct {
	BaseView
	progress   float64 // 0.0 to 1.0
	onChanged  func(progress float64)
	dragging   bool
	hovered    bool
	trackColor color.RGBA
	thumbColor color.RGBA
}

// NewSeekBar creates a new SeekBar with the given initial progress (0.0–1.0).
func NewSeekBar(progress float64) *SeekBar {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	sb := &SeekBar{
		progress:   progress,
		trackColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
		thumbColor: core.ParseColor("#1976D2"),
	}
	sb.node = initNode("SeekBar", sb)
	sb.node.SetPainter(&seekBarPainter{sb: sb})
	sb.node.SetHandler(&seekBarHandler{sb: sb})
	sb.node.SetStyle(&core.Style{})
	return sb
}

// SetProgress sets the current progress (clamped to 0.0–1.0).
func (sb *SeekBar) SetProgress(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	sb.progress = p
	sb.node.MarkDirty()
}

// GetProgress returns the current progress.
func (sb *SeekBar) GetProgress() float64 {
	return sb.progress
}

// SetOnProgressChangedListener sets the callback for progress changes.
func (sb *SeekBar) SetOnProgressChangedListener(fn func(progress float64)) {
	sb.onChanged = fn
}

// SetTrackColor sets the track background color.
func (sb *SeekBar) SetTrackColor(c color.RGBA) {
	sb.trackColor = c
	sb.node.MarkDirty()
}

// SetThumbColor sets the thumb (handle) color.
func (sb *SeekBar) SetThumbColor(c color.RGBA) {
	sb.thumbColor = c
	sb.node.MarkDirty()
}

// updateFromX sets progress based on an x coordinate (global).
func (sb *SeekBar) updateFromX(globalX float64) {
	pos := sb.node.AbsolutePosition()
	b := sb.node.Bounds()
	dpi := getDPIScale(sb.node)
	thumbR := seekBarThumbRadius * dpi
	trackStart := pos.X + thumbR
	trackEnd := pos.X + b.Width - thumbR
	trackLen := trackEnd - trackStart
	if trackLen <= 0 {
		return
	}
	p := (globalX - trackStart) / trackLen
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	sb.progress = p
	sb.node.MarkDirty()
	if sb.onChanged != nil {
		sb.onChanged(p)
	}
}

// ---------- seekBarPainter ----------

type seekBarPainter struct {
	sb *SeekBar
}

func (p *seekBarPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	dpi := getDPIScale(node)
	w := 200.0
	h := seekBarThumbRadius*2*dpi + 8*dpi

	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	} else if ws.Mode == core.MeasureModeAtMost && w > ws.Size {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *seekBarPainter) Paint(node *core.Node, canvas core.Canvas) {
	sb := p.sb
	b := node.Bounds()
	dpi := getDPIScale(node)
	trackH := seekBarTrackHeight * dpi
	thumbR := seekBarThumbRadius * dpi

	centerY := b.Height / 2
	trackStart := thumbR
	trackEnd := b.Width - thumbR
	trackLen := trackEnd - trackStart

	// Track background
	trackRect := core.Rect{
		X:      trackStart,
		Y:      centerY - trackH/2,
		Width:  trackLen,
		Height: trackH,
	}
	trackPaint := &core.Paint{Color: sb.trackColor, DrawStyle: core.PaintFill}
	canvas.DrawRoundRect(trackRect, trackH/2, trackPaint)

	// Active track (filled portion)
	activeW := trackLen * sb.progress
	if activeW > 0 {
		activeRect := core.Rect{
			X:      trackStart,
			Y:      centerY - trackH/2,
			Width:  activeW,
			Height: trackH,
		}
		activePaint := &core.Paint{Color: sb.thumbColor, DrawStyle: core.PaintFill}
		canvas.DrawRoundRect(activeRect, trackH/2, activePaint)
	}

	// Thumb
	thumbX := trackStart + trackLen*sb.progress
	thumbColor := sb.thumbColor
	if sb.dragging {
		// Larger thumb when dragging (no separate halo to avoid aliasing)
		canvas.DrawCircle(thumbX, centerY, thumbR*1.2, &core.Paint{Color: thumbColor, DrawStyle: core.PaintFill})
	} else {
		canvas.DrawCircle(thumbX, centerY, thumbR, &core.Paint{Color: thumbColor, DrawStyle: core.PaintFill})
	}
}

// ---------- seekBarHandler ----------

type seekBarHandler struct {
	core.DefaultHandler
	sb *SeekBar
}

func (h *seekBarHandler) OnEvent(node *core.Node, event core.Event) bool {
	sb := h.sb
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		sb.dragging = true
		sb.updateFromX(me.X)
		return true

	case core.ActionMove:
		if sb.dragging {
			sb.updateFromX(me.X)
			return true
		}

	case core.ActionUp:
		if sb.dragging {
			sb.dragging = false
			sb.updateFromX(me.X)
			return true
		}

	case core.ActionHoverEnter:
		sb.hovered = true
		node.MarkDirty()
		return true

	case core.ActionHoverExit, core.ActionCancel:
		sb.hovered = false
		sb.dragging = false
		node.MarkDirty()
		return true
	}
	return false
}
