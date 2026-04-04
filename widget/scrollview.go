package widget

import (
	"math"
	"time"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
)

// scrollWheelStep is the scroll distance in pixels per mouse wheel notch.
const scrollWheelStep = 48.0

// dragThreshold is the minimum pointer movement in pixels to start a drag.
const dragThreshold = 8.0

// flingVelocityThreshold is the minimum velocity (px/ms) to trigger a fling.
const flingVelocityThreshold = 0.3

// ScrollView is a scrollable container that holds a single child.
// It supports vertical or horizontal scrolling with drag, fling,
// mouse wheel input, and interactive scrollbar.
type ScrollView struct {
	BaseView
	scrollLayout *layout.ScrollLayout
	scrollbar    Scrollbar // shared scrollbar component

	// Touch/mouse scroll tracking
	dragging  bool
	lastDragX float64
	lastDragY float64
	startX    float64
	startY    float64
	velocity  float64
	lastTime  time.Time
	flingAnim *core.ValueAnimator
}

// NewScrollView creates a vertical ScrollView.
func NewScrollView() *ScrollView {
	sv := &ScrollView{}
	sv.scrollLayout = &layout.ScrollLayout{Direction: layout.Vertical}
	sv.scrollbar = Scrollbar{Orientation: layout.Vertical}
	sv.node = initNode("ScrollView", sv)
	sv.node.SetLayout(sv.scrollLayout)
	sv.node.SetPainter(&scrollViewPainter{sv: sv})
	sv.node.SetHandler(&scrollViewHandler{sv: sv})
	sv.node.SetStyle(&core.Style{})
	sv.node.SetData("paintsChildren", true)
	return sv
}

// NewHorizontalScrollView creates a horizontal ScrollView.
func NewHorizontalScrollView() *ScrollView {
	sv := &ScrollView{}
	sv.scrollLayout = &layout.ScrollLayout{Direction: layout.Horizontal}
	sv.scrollbar = Scrollbar{Orientation: layout.Horizontal}
	sv.node = initNode("HorizontalScrollView", sv)
	sv.node.SetLayout(sv.scrollLayout)
	sv.node.SetPainter(&scrollViewPainter{sv: sv})
	sv.node.SetHandler(&scrollViewHandler{sv: sv})
	sv.node.SetStyle(&core.Style{})
	sv.node.SetData("paintsChildren", true)
	return sv
}

// ScrollTo scrolls to the given offset, clamped to valid bounds.
func (sv *ScrollView) ScrollTo(x, y float64) {
	sv.scrollLayout.OffsetX = x
	sv.scrollLayout.OffsetY = y
	sv.clampScroll()
	sv.node.MarkDirty()
}

// GetScrollX returns the current horizontal scroll offset.
func (sv *ScrollView) GetScrollX() float64 {
	return sv.scrollLayout.OffsetX
}

// GetScrollY returns the current vertical scroll offset.
func (sv *ScrollView) GetScrollY() float64 {
	return sv.scrollLayout.OffsetY
}

// Direction returns the scroll orientation.
func (sv *ScrollView) Direction() layout.Orientation {
	return sv.scrollLayout.Direction
}

// clampScroll clamps the scroll offset to valid bounds.
func (sv *ScrollView) clampScroll() {
	viewport := sv.node.MeasuredSize()
	childSize := sv.scrollLayout.ChildSize()
	padding := sv.node.Padding()

	if sv.scrollLayout.Direction == layout.Vertical {
		if sv.scrollLayout.OffsetY < 0 {
			sv.scrollLayout.OffsetY = 0
		}
		maxScroll := childSize.Height - (viewport.Height - padding.Top - padding.Bottom)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if sv.scrollLayout.OffsetY > maxScroll {
			sv.scrollLayout.OffsetY = maxScroll
		}
	} else {
		if sv.scrollLayout.OffsetX < 0 {
			sv.scrollLayout.OffsetX = 0
		}
		maxScroll := childSize.Width - (viewport.Width - padding.Left - padding.Right)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if sv.scrollLayout.OffsetX > maxScroll {
			sv.scrollLayout.OffsetX = maxScroll
		}
	}
}

// startFling begins a fling animation with the given velocity.
func (sv *ScrollView) startFling(velocity float64) {
	distance := velocity * 0.3 // deceleration factor
	var startOffset float64
	if sv.scrollLayout.Direction == layout.Vertical {
		startOffset = sv.scrollLayout.OffsetY
	} else {
		startOffset = sv.scrollLayout.OffsetX
	}

	sv.flingAnim = &core.ValueAnimator{
		From:     0,
		To:       distance,
		Duration: 300 * time.Millisecond,
		Interp:   &core.DecelerateInterpolator{},
		OnUpdate: func(value float64) {
			if sv.scrollLayout.Direction == layout.Vertical {
				sv.scrollLayout.OffsetY = startOffset + value
			} else {
				sv.scrollLayout.OffsetX = startOffset + value
			}
			sv.clampScroll()
			sv.node.MarkDirty()
		},
	}
	sv.flingAnim.Start()
}

// stopFling cancels any running fling animation.
func (sv *ScrollView) stopFling() {
	if sv.flingAnim != nil && sv.flingAnim.IsRunning() {
		sv.flingAnim.Cancel()
	}
	sv.flingAnim = nil
}

// ---------- scrollViewHandler ----------

// scrollViewHandler handles drag, fling, and scroll wheel events.
type scrollViewHandler struct {
	core.DefaultHandler
	sv *ScrollView
}

func (h *scrollViewHandler) OnInterceptEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}
	if me.Action == core.ActionMove {
		sv := h.sv
		if sv.scrollLayout.Direction == layout.Vertical {
			if math.Abs(me.Y-sv.startY) > dragThreshold {
				sv.dragging = true
				sv.lastDragY = me.Y
				sv.lastTime = time.Now()
				return true
			}
		} else {
			if math.Abs(me.X-sv.startX) > dragThreshold {
				sv.dragging = true
				sv.lastDragX = me.X
				sv.lastTime = time.Now()
				return true
			}
		}
	}
	return false
}

func (h *scrollViewHandler) OnEvent(node *core.Node, event core.Event) bool {
	sv := h.sv

	// Handle scroll wheel events
	if se, ok := event.(*core.ScrollEvent); ok {
		if sv.scrollLayout.Direction == layout.Vertical {
			sv.scrollLayout.OffsetY -= se.DeltaY * scrollWheelStep
		} else {
			sv.scrollLayout.OffsetX -= se.DeltaX * scrollWheelStep
		}
		sv.scrollbar.ClearHover()
		sv.clampScroll()
		node.MarkDirty()
		return true
	}

	// Delegate scrollbar interaction to the shared Scrollbar component
	metrics := sv.scrollbarMetrics()
	var currentOffset float64
	if sv.scrollLayout.Direction == layout.Vertical {
		currentOffset = sv.scrollLayout.OffsetY
	} else {
		currentOffset = sv.scrollLayout.OffsetX
	}
	if consumed, newOffset := sv.scrollbar.HandleEvent(node, event, metrics, currentOffset); consumed {
		if sv.scrollbar.Dragging || newOffset != currentOffset {
			if sv.scrollLayout.Direction == layout.Vertical {
				sv.scrollLayout.OffsetY = newOffset
			} else {
				sv.scrollLayout.OffsetX = newOffset
			}
			sv.clampScroll()
		}
		node.MarkDirty()
		return true
	}

	// Handle motion events for content drag scrolling
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		sv.stopFling()
		sv.startX = me.X
		sv.startY = me.Y
		sv.lastDragX = me.X
		sv.lastDragY = me.Y
		sv.lastTime = time.Now()
		sv.velocity = 0
		sv.dragging = false
		return true

	case core.ActionMove:
		if !sv.dragging {
			if sv.scrollLayout.Direction == layout.Vertical {
				if math.Abs(me.Y-sv.startY) > dragThreshold {
					sv.dragging = true
				}
			} else {
				if math.Abs(me.X-sv.startX) > dragThreshold {
					sv.dragging = true
				}
			}
		}
		if sv.dragging {
			now := time.Now()
			dt := now.Sub(sv.lastTime).Seconds() * 1000
			if dt < 1 {
				dt = 1
			}
			if sv.scrollLayout.Direction == layout.Vertical {
				delta := sv.lastDragY - me.Y
				sv.scrollLayout.OffsetY += delta
				sv.velocity = delta / dt
				sv.lastDragY = me.Y
			} else {
				delta := sv.lastDragX - me.X
				sv.scrollLayout.OffsetX += delta
				sv.velocity = delta / dt
				sv.lastDragX = me.X
			}
			sv.lastTime = now
			sv.clampScroll()
			node.MarkDirty()
			return true
		}

	case core.ActionUp:
		if sv.dragging {
			sv.dragging = false
			if math.Abs(sv.velocity) > flingVelocityThreshold {
				sv.startFling(sv.velocity * 1000)
			}
			return true
		}

	case core.ActionHoverExit:
		sv.scrollbar.ClearHover()
		node.MarkDirty()
		return true

	case core.ActionCancel:
		sv.dragging = false
		sv.velocity = 0
		sv.scrollbar.ClearHover()
		node.MarkDirty()
		return true
	}

	return false
}

// ---------- scrollViewPainter ----------

// scrollViewPainter handles measurement and rendering for ScrollView.
// It draws the background, clips to the viewport, translates by scroll
// offset, paints children, and optionally draws a scroll indicator.
type scrollViewPainter struct {
	sv *ScrollView
}

func (p *scrollViewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Delegate to ScrollLayout for measurement
	if l := node.GetLayout(); l != nil {
		return l.Measure(node, ws, hs)
	}
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *scrollViewPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// 1. Draw background if set
	if s != nil && s.BackgroundColor.A > 0 {
		paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}

	// 2. Clip to viewport and paint children with scroll offset
	children := node.Children()
	if len(children) == 0 {
		return
	}

	canvas.Save()
	canvas.ClipRect(localRect)

	// Paint each child with its bounds (which already include scroll offset from Arrange)
	for _, child := range children {
		if child.GetVisibility() != core.Visible {
			continue
		}
		paintNodeRecursive(child, canvas)
	}

	canvas.Restore()

	// 3. Draw scroll indicator
	p.paintScrollIndicator(node, canvas)
}

// paintNodeRecursive paints a node and its descendants.
// This mirrors the standard PaintNode logic used in app/ and platform/.
func paintNodeRecursive(node *core.Node, canvas core.Canvas) {
	if node.GetVisibility() != core.Visible {
		return
	}
	canvas.Save()
	b := node.Bounds()
	canvas.Translate(b.X, b.Y)
	if painter := node.GetPainter(); painter != nil {
		painter.Paint(node, canvas)
	}
	// If this child also paints its own children, skip recursive child painting
	if node.GetData("paintsChildren") == nil {
		for _, child := range node.Children() {
			paintNodeRecursive(child, canvas)
		}
	}
	canvas.Restore()
}

// scrollbarMetrics returns metrics for the shared Scrollbar component.
func (sv *ScrollView) scrollbarMetrics() ScrollbarMetrics {
	b := sv.node.Bounds()
	childSize := sv.scrollLayout.ChildSize()
	if sv.scrollLayout.Direction == layout.Vertical {
		return sv.scrollbar.ComputeMetrics(b.Height, childSize.Height, sv.scrollLayout.OffsetY)
	}
	return sv.scrollbar.ComputeMetrics(b.Width, childSize.Width, sv.scrollLayout.OffsetX)
}

// paintScrollIndicator delegates to the shared Scrollbar component.
func (p *scrollViewPainter) paintScrollIndicator(node *core.Node, canvas core.Canvas) {
	sv := p.sv
	b := node.Bounds()
	metrics := sv.scrollbarMetrics()
	sv.scrollbar.Paint(canvas, core.Rect{Width: b.Width, Height: b.Height}, metrics)
}
