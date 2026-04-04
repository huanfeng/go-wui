package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
)

const (
	splitDividerSize = 6.0 // divider thickness in dp
	splitMinPaneSize = 30.0
)

// SplitPane divides its area into two resizable panes with a draggable divider.
// Supports horizontal (left/right) and vertical (top/bottom) orientation.
type SplitPane struct {
	BaseView
	orientation layout.Orientation
	ratio       float64 // 0.0–1.0, position of divider
	firstPane   core.View
	secondPane  core.View
	dragging    bool
	hovered     bool // divider hovered
}

// NewSplitPane creates a new SplitPane with the given orientation.
// ratio (0.0–1.0) sets the initial divider position.
func NewSplitPane(orientation layout.Orientation, ratio float64) *SplitPane {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	sp := &SplitPane{
		orientation: orientation,
		ratio:       ratio,
	}
	sp.node = initNode("SplitPane", sp)
	sp.node.SetPainter(&splitPanePainter{sp: sp})
	sp.node.SetHandler(&splitPaneHandler{sp: sp})
	sp.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 245, G: 245, B: 245, A: 255},
	})
	sp.node.SetData("paintsChildren", true)
	return sp
}

// SetFirstPane sets the view for the first pane (left or top).
func (sp *SplitPane) SetFirstPane(v core.View) {
	if sp.firstPane != nil {
		sp.node.RemoveChild(sp.firstPane.Node())
	}
	sp.firstPane = v
	if v != nil {
		sp.node.AddChild(v.Node())
	}
	sp.node.MarkDirty()
}

// SetSecondPane sets the view for the second pane (right or bottom).
func (sp *SplitPane) SetSecondPane(v core.View) {
	if sp.secondPane != nil {
		sp.node.RemoveChild(sp.secondPane.Node())
	}
	sp.secondPane = v
	if v != nil {
		sp.node.AddChild(v.Node())
	}
	sp.node.MarkDirty()
}

// GetFirstPane returns the first pane view.
func (sp *SplitPane) GetFirstPane() core.View {
	return sp.firstPane
}

// GetSecondPane returns the second pane view.
func (sp *SplitPane) GetSecondPane() core.View {
	return sp.secondPane
}

// SetRatio sets the divider position (0.0–1.0).
func (sp *SplitPane) SetRatio(ratio float64) {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	sp.ratio = ratio
	sp.node.MarkDirty()
}

// GetRatio returns the current divider position.
func (sp *SplitPane) GetRatio() float64 {
	return sp.ratio
}

// GetOrientation returns the split orientation.
func (sp *SplitPane) GetOrientation() layout.Orientation {
	return sp.orientation
}

// dividerRect returns the divider's local rectangle.
func (sp *SplitPane) dividerRect(bounds core.Rect, dpi float64) core.Rect {
	divSize := splitDividerSize * dpi
	if sp.orientation == layout.Horizontal {
		divX := bounds.Width * sp.ratio
		return core.Rect{X: divX - divSize/2, Y: 0, Width: divSize, Height: bounds.Height}
	}
	divY := bounds.Height * sp.ratio
	return core.Rect{X: 0, Y: divY - divSize/2, Width: bounds.Width, Height: divSize}
}

// ---------- splitPanePainter ----------

type splitPanePainter struct {
	sp *SplitPane
}

func (p *splitPanePainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *splitPanePainter) Paint(node *core.Node, canvas core.Canvas) {
	sp := p.sp
	s := node.GetStyle()
	b := node.Bounds()
	dpi := getDPIScale(node)
	divSize := splitDividerSize * dpi

	// Background
	if s != nil && s.BackgroundColor.A > 0 {
		bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, bgPaint)
	}

	// Calculate pane bounds
	var firstBounds, secondBounds core.Rect
	if sp.orientation == layout.Horizontal {
		divX := b.Width * sp.ratio
		firstBounds = core.Rect{X: 0, Y: 0, Width: divX - divSize/2, Height: b.Height}
		secondBounds = core.Rect{X: divX + divSize/2, Y: 0, Width: b.Width - divX - divSize/2, Height: b.Height}
	} else {
		divY := b.Height * sp.ratio
		firstBounds = core.Rect{X: 0, Y: 0, Width: b.Width, Height: divY - divSize/2}
		secondBounds = core.Rect{X: 0, Y: divY + divSize/2, Width: b.Width, Height: b.Height - divY - divSize/2}
	}

	// Measure and paint first pane
	if sp.firstPane != nil {
		pn := sp.firstPane.Node()
		pn.SetBounds(firstBounds)
		pn.SetMeasuredSize(core.Size{Width: firstBounds.Width, Height: firstBounds.Height})
		if l := pn.GetLayout(); l != nil {
			ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: firstBounds.Width}
			hs := core.MeasureSpec{Mode: core.MeasureModeExact, Size: firstBounds.Height}
			l.Measure(pn, ws, hs)
			l.Arrange(pn, firstBounds)
		}
		paintNodeRecursive(pn, canvas)
	}

	// Measure and paint second pane
	if sp.secondPane != nil {
		pn := sp.secondPane.Node()
		pn.SetBounds(secondBounds)
		pn.SetMeasuredSize(core.Size{Width: secondBounds.Width, Height: secondBounds.Height})
		if l := pn.GetLayout(); l != nil {
			ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: secondBounds.Width}
			hs := core.MeasureSpec{Mode: core.MeasureModeExact, Size: secondBounds.Height}
			l.Measure(pn, ws, hs)
			l.Arrange(pn, secondBounds)
		}
		paintNodeRecursive(pn, canvas)
	}

	// Draw divider
	divRect := sp.dividerRect(core.Rect{Width: b.Width, Height: b.Height}, dpi)
	divColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	if sp.dragging {
		divColor = color.RGBA{R: 25, G: 118, B: 210, A: 255}
	} else if sp.hovered {
		divColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
	}
	divPaint := &core.Paint{Color: divColor, DrawStyle: core.PaintFill}
	canvas.DrawRect(divRect, divPaint)

	// Draw grip dots on divider
	gripColor := color.RGBA{R: 160, G: 160, B: 160, A: 255}
	if sp.dragging || sp.hovered {
		gripColor = color.RGBA{R: 255, G: 255, B: 255, A: 200}
	}
	gripPaint := &core.Paint{Color: gripColor, DrawStyle: core.PaintFill}
	gripR := 1.5 * dpi
	if sp.orientation == layout.Horizontal {
		cx := divRect.X + divRect.Width/2
		cy := b.Height / 2
		canvas.DrawCircle(cx, cy-8*dpi, gripR, gripPaint)
		canvas.DrawCircle(cx, cy, gripR, gripPaint)
		canvas.DrawCircle(cx, cy+8*dpi, gripR, gripPaint)
	} else {
		cx := b.Width / 2
		cy := divRect.Y + divRect.Height/2
		canvas.DrawCircle(cx-8*dpi, cy, gripR, gripPaint)
		canvas.DrawCircle(cx, cy, gripR, gripPaint)
		canvas.DrawCircle(cx+8*dpi, cy, gripR, gripPaint)
	}
}

// ---------- splitPaneHandler ----------

type splitPaneHandler struct {
	core.DefaultHandler
	sp *SplitPane
}

func (h *splitPaneHandler) isDividerHit(globalX, globalY float64) bool {
	sp := h.sp
	pos := sp.node.AbsolutePosition()
	localX := globalX - pos.X
	localY := globalY - pos.Y
	dpi := getDPIScale(sp.node)
	b := sp.node.Bounds()
	dr := sp.dividerRect(core.Rect{Width: b.Width, Height: b.Height}, dpi)
	// Expand hit area slightly for easier grabbing
	expand := 4.0 * dpi
	return localX >= dr.X-expand && localX <= dr.X+dr.Width+expand &&
		localY >= dr.Y-expand && localY <= dr.Y+dr.Height+expand
}

func (h *splitPaneHandler) updateRatio(globalX, globalY float64) {
	sp := h.sp
	pos := sp.node.AbsolutePosition()
	b := sp.node.Bounds()

	var ratio float64
	if sp.orientation == layout.Horizontal {
		localX := globalX - pos.X
		ratio = localX / b.Width
	} else {
		localY := globalY - pos.Y
		ratio = localY / b.Height
	}

	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	sp.ratio = ratio
	sp.node.MarkDirty()
}

func (h *splitPaneHandler) OnEvent(node *core.Node, event core.Event) bool {
	sp := h.sp
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		if h.isDividerHit(me.X, me.Y) {
			sp.dragging = true
			node.MarkDirty()
			return true
		}

	case core.ActionMove:
		if sp.dragging {
			h.updateRatio(me.X, me.Y)
			return true
		}

	case core.ActionUp:
		if sp.dragging {
			sp.dragging = false
			h.updateRatio(me.X, me.Y)
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		wasHovered := sp.hovered
		sp.hovered = h.isDividerHit(me.X, me.Y)
		if sp.hovered != wasHovered {
			node.MarkDirty()
		}
		return sp.hovered

	case core.ActionHoverExit, core.ActionCancel:
		sp.hovered = false
		sp.dragging = false
		node.MarkDirty()
		return true
	}
	return false
}
