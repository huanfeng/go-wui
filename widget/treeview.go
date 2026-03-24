package widget

import (
	"image/color"

	"github.com/huanfeng/go-wui/core"
)

const (
	treeNodeHeight = 32.0
	treeIndent     = 20.0
)

// TreeNode represents a node in a TreeView.
type TreeNode struct {
	Text     string
	Children []*TreeNode
	Expanded bool
	Data     interface{} // user data
}

// AddChild appends a child node.
func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

// IsLeaf reports whether this node has no children.
func (n *TreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// flatEntry is a flattened tree node with its depth for rendering.
type flatEntry struct {
	node  *TreeNode
	depth int
}

// TreeView displays a hierarchical tree with expand/collapse support.
// Each node can have children, and non-leaf nodes show an expand/collapse toggle.
type TreeView struct {
	BaseView
	roots       []*TreeNode
	selectedIdx int // index in the flattened list
	onSelected  func(node *TreeNode)
	flat        []flatEntry // cached flattened view

	hoveredIdx int
	pressedIdx int
	scrollY    float64
}

// NewTreeView creates a new TreeView.
func NewTreeView() *TreeView {
	tv := &TreeView{
		selectedIdx: -1,
		hoveredIdx:  -1,
		pressedIdx:  -1,
	}
	tv.node = initNode("TreeView", tv)
	tv.node.SetPainter(&treeViewPainter{tv: tv})
	tv.node.SetHandler(&treeViewHandler{tv: tv})
	tv.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 33, G: 33, B: 33, A: 255},
		FontSize:        14,
	})
	tv.node.SetData("paintsChildren", true)
	return tv
}

// SetRoots sets the root-level tree nodes.
func (tv *TreeView) SetRoots(roots []*TreeNode) {
	tv.roots = roots
	tv.rebuildFlat()
	tv.node.MarkDirty()
}

// GetRoots returns the root-level tree nodes.
func (tv *TreeView) GetRoots() []*TreeNode {
	return tv.roots
}

// AddRoot appends a root-level node.
func (tv *TreeView) AddRoot(node *TreeNode) {
	tv.roots = append(tv.roots, node)
	tv.rebuildFlat()
	tv.node.MarkDirty()
}

// SetOnNodeSelectedListener sets the callback when a node is selected.
func (tv *TreeView) SetOnNodeSelectedListener(fn func(node *TreeNode)) {
	tv.onSelected = fn
}

// GetSelectedNode returns the currently selected tree node, or nil.
func (tv *TreeView) GetSelectedNode() *TreeNode {
	if tv.selectedIdx >= 0 && tv.selectedIdx < len(tv.flat) {
		return tv.flat[tv.selectedIdx].node
	}
	return nil
}

// GetFlatCount returns the number of visible (flattened) entries.
func (tv *TreeView) GetFlatCount() int {
	return len(tv.flat)
}

// rebuildFlat flattens the visible tree into a list.
func (tv *TreeView) rebuildFlat() {
	tv.flat = nil
	for _, root := range tv.roots {
		tv.flattenNode(root, 0)
	}
}

func (tv *TreeView) flattenNode(node *TreeNode, depth int) {
	tv.flat = append(tv.flat, flatEntry{node: node, depth: depth})
	if node.Expanded {
		for _, child := range node.Children {
			tv.flattenNode(child, depth+1)
		}
	}
}

// ---------- treeViewPainter ----------

type treeViewPainter struct {
	tv *TreeView
}

func (p *treeViewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *treeViewPainter) Paint(node *core.Node, canvas core.Canvas) {
	tv := p.tv
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	dpi := getDPIScale(node)
	itemH := treeNodeHeight * dpi
	indent := treeIndent * dpi

	// Background
	if s.BackgroundColor.A > 0 {
		bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, bgPaint)
	}

	canvas.Save()
	canvas.ClipRect(core.Rect{Width: b.Width, Height: b.Height})

	fontSize := s.FontSize * dpi
	if fontSize == 0 {
		fontSize = 14 * dpi
	}

	for i, entry := range tv.flat {
		itemY := float64(i)*itemH - tv.scrollY
		if itemY+itemH < 0 || itemY > b.Height {
			continue // outside viewport
		}

		itemRect := core.Rect{X: 0, Y: itemY, Width: b.Width, Height: itemH}

		// Selection highlight
		if i == tv.selectedIdx {
			selPaint := &core.Paint{Color: color.RGBA{R: 25, G: 118, B: 210, A: 30}, DrawStyle: core.PaintFill}
			canvas.DrawRect(itemRect, selPaint)
		} else if i == tv.pressedIdx {
			hlPaint := &core.Paint{Color: color.RGBA{A: 20}, DrawStyle: core.PaintFill}
			canvas.DrawRect(itemRect, hlPaint)
		} else if i == tv.hoveredIdx {
			hlPaint := &core.Paint{Color: color.RGBA{A: 10}, DrawStyle: core.PaintFill}
			canvas.DrawRect(itemRect, hlPaint)
		}

		xOffset := 8*dpi + float64(entry.depth)*indent

		// Expand/collapse toggle for non-leaf nodes
		if !entry.node.IsLeaf() {
			togglePaint := &core.Paint{Color: s.TextColor, FontSize: fontSize * 0.8}
			toggle := "\u25B6" // ▶
			if entry.node.Expanded {
				toggle = "\u25BC" // ▼
			}
			toggleSize := canvas.MeasureText(toggle, togglePaint)
			toggleY := itemY + (itemH-toggleSize.Height)/2
			canvas.DrawText(toggle, xOffset, toggleY, togglePaint)
			xOffset += toggleSize.Width + 6*dpi
		} else {
			xOffset += fontSize*0.8 + 6*dpi // same indent as toggle width
		}

		// Node text
		textPaint := &core.Paint{Color: s.TextColor, FontSize: fontSize}
		textSize := canvas.MeasureText(entry.node.Text, textPaint)
		textY := itemY + (itemH-textSize.Height)/2
		canvas.DrawText(entry.node.Text, xOffset, textY, textPaint)
	}

	canvas.Restore()

	// Scroll indicator
	totalH := float64(len(tv.flat)) * itemH
	if totalH > b.Height {
		ratio := b.Height / totalH
		indicatorH := b.Height * ratio
		if indicatorH < 20 {
			indicatorH = 20
		}
		scrollRange := totalH - b.Height
		fraction := tv.scrollY / scrollRange
		trackH := b.Height - indicatorH
		indicatorY := trackH * fraction

		indicatorPaint := &core.Paint{Color: color.RGBA{R: 128, G: 128, B: 128, A: 80}, DrawStyle: core.PaintFill}
		canvas.DrawRoundRect(core.Rect{
			X: b.Width - 4*dpi - 2*dpi, Y: indicatorY,
			Width: 4 * dpi, Height: indicatorH,
		}, 2*dpi, indicatorPaint)
	}
}

// ---------- treeViewHandler ----------

type treeViewHandler struct {
	core.DefaultHandler
	tv *TreeView
}

func (h *treeViewHandler) positionAtY(globalY float64) int {
	tv := h.tv
	pos := tv.node.AbsolutePosition()
	localY := globalY - pos.Y
	dpi := getDPIScale(tv.node)
	itemH := treeNodeHeight * dpi
	idx := int((localY + tv.scrollY) / itemH)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(tv.flat) {
		idx = len(tv.flat) - 1
	}
	return idx
}

func (h *treeViewHandler) OnEvent(node *core.Node, event core.Event) bool {
	tv := h.tv

	// Scroll wheel
	if se, ok := event.(*core.ScrollEvent); ok {
		dpi := getDPIScale(node)
		tv.scrollY -= se.DeltaY * 48 * dpi
		h.clampScroll()
		node.MarkDirty()
		return true
	}

	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		idx := h.positionAtY(me.Y)
		if idx >= 0 && idx < len(tv.flat) {
			tv.pressedIdx = idx
			node.MarkDirty()
		}
		return true

	case core.ActionUp:
		if tv.pressedIdx >= 0 {
			idx := h.positionAtY(me.Y)
			pressed := tv.pressedIdx
			tv.pressedIdx = -1
			node.MarkDirty()

			if idx == pressed && idx >= 0 && idx < len(tv.flat) {
				entry := tv.flat[idx]
				// Toggle expand/collapse for non-leaf
				if !entry.node.IsLeaf() {
					entry.node.Expanded = !entry.node.Expanded
					tv.rebuildFlat()
				}
				tv.selectedIdx = idx
				if tv.onSelected != nil {
					tv.onSelected(entry.node)
				}
			}
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		idx := h.positionAtY(me.Y)
		if idx != tv.hoveredIdx {
			tv.hoveredIdx = idx
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		tv.hoveredIdx = -1
		tv.pressedIdx = -1
		node.MarkDirty()
		return true
	}
	return false
}

func (h *treeViewHandler) clampScroll() {
	tv := h.tv
	dpi := getDPIScale(tv.node)
	itemH := treeNodeHeight * dpi
	totalH := float64(len(tv.flat)) * itemH
	viewportH := tv.node.Bounds().Height
	if viewportH == 0 {
		viewportH = tv.node.MeasuredSize().Height
	}
	maxScroll := totalH - viewportH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tv.scrollY < 0 {
		tv.scrollY = 0
	}
	if tv.scrollY > maxScroll {
		tv.scrollY = maxScroll
	}
}
