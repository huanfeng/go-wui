package widget

import (
	"image/color"
	"math"
	"time"

	"gowui/core"
	"gowui/layout"
)

// RecyclerAdapter provides data and creates ViewHolders for RecyclerView.
// Modeled after Android's RecyclerView.Adapter.
type RecyclerAdapter interface {
	// GetItemCount returns the total number of items.
	GetItemCount() int
	// GetItemViewType returns the view type for the item at the given position.
	// Default implementation should return 0 for single type.
	GetItemViewType(position int) int
	// CreateViewHolder creates a new ViewHolder for the given view type.
	CreateViewHolder(viewType int) *ViewHolder
	// BindViewHolder binds data to the ViewHolder at the given position.
	BindViewHolder(holder *ViewHolder, position int)
}

// ViewHolder holds a reference to the item view and its position.
type ViewHolder struct {
	ItemView core.View
	position int
	viewType int
	bound    bool
}

// GetPosition returns the adapter position this holder is bound to.
func (vh *ViewHolder) GetPosition() int {
	return vh.position
}

// recyclerPool stores recycled ViewHolders by view type.
type recyclerPool struct {
	pool map[int][]*ViewHolder
}

func newRecyclerPool() *recyclerPool {
	return &recyclerPool{pool: make(map[int][]*ViewHolder)}
}

func (p *recyclerPool) get(viewType int) *ViewHolder {
	holders := p.pool[viewType]
	if len(holders) == 0 {
		return nil
	}
	vh := holders[len(holders)-1]
	p.pool[viewType] = holders[:len(holders)-1]
	return vh
}

func (p *recyclerPool) put(vh *ViewHolder) {
	vh.bound = false
	p.pool[vh.viewType] = append(p.pool[vh.viewType], vh)
}

func (p *recyclerPool) clear() {
	p.pool = make(map[int][]*ViewHolder)
}

// RecyclerView is a high-performance list that recycles item views.
// It uses an Adapter pattern and a recycle pool to minimize view creation.
type RecyclerView struct {
	BaseView
	adapter     RecyclerAdapter
	pool        *recyclerPool
	scrollY     float64
	itemHeight  float64 // fixed item height for simplicity
	activeHolders map[int]*ViewHolder
	scrollbar   Scrollbar // shared scrollbar component // position → holder

	// Scroll tracking (reuses ScrollView patterns)
	dragging  bool
	lastDragY float64
	startY    float64
	velocity  float64
	lastTime  time.Time
	flingAnim *core.ValueAnimator

	onItemClick func(position int)
	hoveredPos  int
	pressedPos  int
}

// NewRecyclerView creates a new RecyclerView with the given item height.
func NewRecyclerView(itemHeight float64) *RecyclerView {
	if itemHeight <= 0 {
		itemHeight = 48
	}
	rv := &RecyclerView{
		pool:          newRecyclerPool(),
		itemHeight:    itemHeight,
		activeHolders: make(map[int]*ViewHolder),
		scrollbar:     Scrollbar{Orientation: layout.Vertical},
		hoveredPos:    -1,
		pressedPos:    -1,
	}
	rv.node = initNode("RecyclerView", rv)
	rv.node.SetPainter(&recyclerViewPainter{rv: rv})
	rv.node.SetHandler(&recyclerViewHandler{rv: rv})
	rv.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	rv.node.SetData("paintsChildren", true)
	return rv
}

// SetAdapter sets the adapter and rebuilds the view.
func (rv *RecyclerView) SetAdapter(adapter RecyclerAdapter) {
	rv.adapter = adapter
	rv.scrollY = 0
	rv.pool.clear()
	rv.activeHolders = make(map[int]*ViewHolder)
	// Remove old children
	for _, child := range rv.node.Children() {
		rv.node.RemoveChild(child)
	}
	rv.node.MarkDirty()
}

// GetAdapter returns the current adapter.
func (rv *RecyclerView) GetAdapter() RecyclerAdapter {
	return rv.adapter
}

// SetItemHeight sets the fixed height for each item.
func (rv *RecyclerView) SetItemHeight(h float64) {
	rv.itemHeight = h
	rv.node.MarkDirty()
}

// GetItemHeight returns the fixed item height.
func (rv *RecyclerView) GetItemHeight() float64 {
	return rv.itemHeight
}

// ScrollToPosition scrolls to make the given position visible.
func (rv *RecyclerView) ScrollToPosition(position int) {
	rv.scrollY = float64(position) * rv.itemHeight
	rv.clampScroll()
	rv.node.MarkDirty()
}

// GetScrollY returns the current vertical scroll offset.
func (rv *RecyclerView) GetScrollY() float64 {
	return rv.scrollY
}

// SetOnItemClickListener sets the click handler for item clicks.
func (rv *RecyclerView) SetOnItemClickListener(fn func(position int)) {
	rv.onItemClick = fn
}

// NotifyDataSetChanged signals the adapter data has changed.
func (rv *RecyclerView) NotifyDataSetChanged() {
	rv.pool.clear()
	rv.activeHolders = make(map[int]*ViewHolder)
	for _, child := range rv.node.Children() {
		rv.node.RemoveChild(child)
	}
	rv.clampScroll()
	rv.node.MarkDirty()
}

// totalContentHeight returns the total scrollable content height.
func (rv *RecyclerView) totalContentHeight() float64 {
	if rv.adapter == nil {
		return 0
	}
	return float64(rv.adapter.GetItemCount()) * rv.itemHeight
}

// clampScroll clamps scroll offset to valid range.
func (rv *RecyclerView) clampScroll() {
	if rv.scrollY < 0 {
		rv.scrollY = 0
	}
	viewportH := rv.node.MeasuredSize().Height
	if viewportH == 0 {
		viewportH = rv.node.Bounds().Height
	}
	maxScroll := rv.totalContentHeight() - viewportH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if rv.scrollY > maxScroll {
		rv.scrollY = maxScroll
	}
}

// startFling begins a fling animation.
func (rv *RecyclerView) startFling(velocity float64) {
	distance := velocity * 0.3
	startOffset := rv.scrollY

	rv.flingAnim = &core.ValueAnimator{
		From:     0,
		To:       distance,
		Duration: 300 * time.Millisecond,
		Interp:   &core.DecelerateInterpolator{},
		OnUpdate: func(value float64) {
			rv.scrollY = startOffset + value
			rv.clampScroll()
			rv.node.MarkDirty()
		},
	}
	rv.flingAnim.Start()
}

func (rv *RecyclerView) stopFling() {
	if rv.flingAnim != nil && rv.flingAnim.IsRunning() {
		rv.flingAnim.Cancel()
	}
	rv.flingAnim = nil
}

// rvScrollbarMetrics returns metrics for the shared Scrollbar component.
func (rv *RecyclerView) rvScrollbarMetrics() ScrollbarMetrics {
	b := rv.node.Bounds()
	if b.Height == 0 {
		b.Height = rv.node.MeasuredSize().Height
	}
	return rv.scrollbar.ComputeMetrics(b.Height, rv.totalContentHeight(), rv.scrollY)
}

// positionAtY returns the item position at the given y coordinate (global/window coords).
func (rv *RecyclerView) positionAtY(globalY float64) int {
	// Convert global coordinate to local (relative to RecyclerView)
	absPos := rv.node.AbsolutePosition()
	localY := globalY - absPos.Y
	pos := int((localY + rv.scrollY) / rv.itemHeight)
	if rv.adapter != nil && pos >= rv.adapter.GetItemCount() {
		pos = rv.adapter.GetItemCount() - 1
	}
	if pos < 0 {
		pos = 0
	}
	return pos
}

// ---------- recyclerViewPainter ----------

type recyclerViewPainter struct {
	rv *RecyclerView
}

func (p *recyclerViewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *recyclerViewPainter) Paint(node *core.Node, canvas core.Canvas) {
	rv := p.rv
	s := node.GetStyle()
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// Background
	if s != nil && s.BackgroundColor.A > 0 {
		bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		canvas.DrawRect(localRect, bgPaint)
	}

	if rv.adapter == nil || rv.adapter.GetItemCount() == 0 {
		return
	}

	canvas.Save()
	canvas.ClipRect(localRect)

	// Calculate visible range
	firstVisible := int(rv.scrollY / rv.itemHeight)
	if firstVisible < 0 {
		firstVisible = 0
	}
	visibleCount := int(b.Height/rv.itemHeight) + 2 // +2 for partial items
	lastVisible := firstVisible + visibleCount
	itemCount := rv.adapter.GetItemCount()
	if lastVisible > itemCount {
		lastVisible = itemCount
	}

	// Recycle holders that are no longer visible
	newActive := make(map[int]*ViewHolder)
	for pos, holder := range rv.activeHolders {
		if pos < firstVisible || pos >= lastVisible {
			rv.pool.put(holder)
		} else {
			newActive[pos] = holder
		}
	}
	rv.activeHolders = newActive

	// Bind and paint visible items
	for pos := firstVisible; pos < lastVisible; pos++ {
		holder, exists := rv.activeHolders[pos]
		if !exists {
			viewType := rv.adapter.GetItemViewType(pos)
			holder = rv.pool.get(viewType)
			if holder == nil {
				holder = rv.adapter.CreateViewHolder(viewType)
				holder.viewType = viewType
			}
			holder.position = pos
			rv.adapter.BindViewHolder(holder, pos)
			holder.bound = true
			rv.activeHolders[pos] = holder
		}

		// Position the item
		itemY := float64(pos)*rv.itemHeight - rv.scrollY
		itemNode := holder.ItemView.Node()
		itemNode.SetBounds(core.Rect{X: 0, Y: itemY, Width: b.Width, Height: rv.itemHeight})
		itemNode.SetMeasuredSize(core.Size{Width: b.Width, Height: rv.itemHeight})

		// Hover/press highlight
		if rv.pressedPos == pos {
			hlPaint := &core.Paint{Color: color.RGBA{A: 25}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: 0, Y: itemY, Width: b.Width, Height: rv.itemHeight}, hlPaint)
		} else if rv.hoveredPos == pos {
			hlPaint := &core.Paint{Color: color.RGBA{A: 12}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: 0, Y: itemY, Width: b.Width, Height: rv.itemHeight}, hlPaint)
		}

		// Paint item
		paintNodeRecursive(itemNode, canvas)
	}

	canvas.Restore()

	// Scroll indicator
	p.paintScrollIndicator(node, canvas)
}

func (p *recyclerViewPainter) paintScrollIndicator(node *core.Node, canvas core.Canvas) {
	rv := p.rv
	b := node.Bounds()
	metrics := rv.rvScrollbarMetrics()
	rv.scrollbar.Paint(canvas, core.Rect{Width: b.Width, Height: b.Height}, metrics)
}

// ---------- recyclerViewHandler ----------

const rvScrollWheelStep = 48.0
const rvDragThreshold = 8.0
const rvFlingVelocityThreshold = 0.3

type recyclerViewHandler struct {
	core.DefaultHandler
	rv *RecyclerView
}

func (h *recyclerViewHandler) OnInterceptEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}
	if me.Action == core.ActionMove {
		rv := h.rv
		if math.Abs(me.Y-rv.startY) > rvDragThreshold {
			rv.dragging = true
			rv.lastDragY = me.Y
			rv.lastTime = time.Now()
			return true
		}
	}
	return false
}

func (h *recyclerViewHandler) OnEvent(node *core.Node, event core.Event) bool {
	rv := h.rv

	// Scroll wheel
	if se, ok := event.(*core.ScrollEvent); ok {
		rv.scrollY -= se.DeltaY * rvScrollWheelStep
		rv.scrollbar.ClearHover()
		rv.clampScroll()
		node.MarkDirty()
		return true
	}

	// Delegate scrollbar interaction first (before list item handling)
	metrics := rv.rvScrollbarMetrics()
	if consumed, newOffset := rv.scrollbar.HandleEvent(node, event, metrics, rv.scrollY); consumed {
		if rv.scrollbar.Dragging || newOffset != rv.scrollY {
			rv.scrollY = newOffset
			rv.clampScroll()
		}
		node.MarkDirty()
		return true
	}

	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		rv.stopFling()
		rv.startY = me.Y
		rv.lastDragY = me.Y
		rv.lastTime = time.Now()
		rv.velocity = 0
		rv.dragging = false
		rv.pressedPos = rv.positionAtY(me.Y)
		node.MarkDirty()
		return true

	case core.ActionMove:
		if !rv.dragging {
			if math.Abs(me.Y-rv.startY) > rvDragThreshold {
				rv.dragging = true
				rv.pressedPos = -1 // cancel press when dragging
			}
		}
		if rv.dragging {
			now := time.Now()
			dt := now.Sub(rv.lastTime).Seconds() * 1000
			if dt < 1 {
				dt = 1
			}
			delta := rv.lastDragY - me.Y
			rv.scrollY += delta
			rv.velocity = delta / dt
			rv.lastDragY = me.Y
			rv.lastTime = now
			rv.clampScroll()
			node.MarkDirty()
			return true
		}

	case core.ActionUp:
		if rv.dragging {
			rv.dragging = false
			if math.Abs(rv.velocity) > rvFlingVelocityThreshold {
				rv.startFling(rv.velocity * 1000)
			}
			return true
		}
		// Click on item
		if rv.pressedPos >= 0 {
			pos := rv.pressedPos
			rv.pressedPos = -1
			node.MarkDirty()
			hit := rv.positionAtY(me.Y)
			if hit == pos && rv.onItemClick != nil && node.IsEnabled() {
				rv.onItemClick(pos)
			}
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		pos := rv.positionAtY(me.Y)
		if pos != rv.hoveredPos {
			rv.hoveredPos = pos
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		rv.hoveredPos = -1
		rv.pressedPos = -1
		rv.scrollbar.ClearHover()
		rv.dragging = false
		node.MarkDirty()
		return true
	}

	return false
}
