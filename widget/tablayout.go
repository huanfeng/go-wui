package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
)

// defaultTabHeight is the standard tab bar height in dp.
const defaultTabHeight = 48.0

// indicatorHeight is the height of the selected tab indicator.
const indicatorHeight = 3.0

// Tab represents a single tab in a TabLayout.
type Tab struct {
	Text string
}

// TabLayout is a horizontal strip of tabs with a selected indicator.
// Modeled after Android's TabLayout.
type TabLayout struct {
	BaseView
	tabs        []Tab
	selectedTab int
	onSelected  func(index int)

	// interaction tracking
	hoveredTab int // -1 = none
	pressedTab int // -1 = none
}

// NewTabLayout creates a new TabLayout.
func NewTabLayout() *TabLayout {
	tl := &TabLayout{
		selectedTab: 0,
		hoveredTab:  -1,
		pressedTab:  -1,
	}
	tl.node = initNode("TabLayout", tl)
	tl.node.SetPainter(&tabLayoutPainter{tl: tl})
	tl.node.SetHandler(&tabLayoutHandler{tl: tl})
	tl.node.SetStyle(&core.Style{
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		FontSize:        14,
	})
	return tl
}

// AddTab appends a tab.
func (tl *TabLayout) AddTab(tab Tab) {
	tl.tabs = append(tl.tabs, tab)
	tl.node.MarkDirty()
}

// RemoveTabAt removes the tab at the given index.
func (tl *TabLayout) RemoveTabAt(index int) {
	if index < 0 || index >= len(tl.tabs) {
		return
	}
	tl.tabs = append(tl.tabs[:index], tl.tabs[index+1:]...)
	if tl.selectedTab >= len(tl.tabs) && len(tl.tabs) > 0 {
		tl.selectedTab = len(tl.tabs) - 1
	}
	tl.node.MarkDirty()
}

// GetTabCount returns the number of tabs.
func (tl *TabLayout) GetTabCount() int {
	return len(tl.tabs)
}

// SetSelectedTab sets the selected tab index.
func (tl *TabLayout) SetSelectedTab(index int) {
	if index < 0 || index >= len(tl.tabs) {
		return
	}
	if tl.selectedTab == index {
		return
	}
	tl.selectedTab = index
	tl.node.MarkDirty()
	if tl.onSelected != nil {
		tl.onSelected(index)
	}
}

// GetSelectedTab returns the currently selected tab index.
func (tl *TabLayout) GetSelectedTab() int {
	return tl.selectedTab
}

// SetOnTabSelectedListener sets the callback invoked when a tab is selected.
func (tl *TabLayout) SetOnTabSelectedListener(fn func(index int)) {
	tl.onSelected = fn
}

// tabWidth returns the width of each tab (equal distribution).
func (tl *TabLayout) tabWidth() float64 {
	if len(tl.tabs) == 0 {
		return 0
	}
	b := tl.node.Bounds()
	return b.Width / float64(len(tl.tabs))
}

// ---------- tabLayoutPainter ----------

type tabLayoutPainter struct {
	tl *TabLayout
}

func (p *tabLayoutPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	h := defaultTabHeight
	w := 300.0

	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	} else if ws.Mode == core.MeasureModeAtMost {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *tabLayoutPainter) Paint(node *core.Node, canvas core.Canvas) {
	tl := p.tl
	s := node.GetStyle()
	if s == nil || len(tl.tabs) == 0 {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// 1. Background
	bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
	canvas.DrawRect(localRect, bgPaint)

	tabW := b.Width / float64(len(tl.tabs))
	fontSize := s.FontSize
	if fontSize == 0 {
		fontSize = 14
	}
	textColor := s.TextColor
	if textColor.A == 0 {
		textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	// 2. Draw each tab
	for i, tab := range tl.tabs {
		tabX := float64(i) * tabW

		// Hover/press highlight
		if tl.pressedTab == i {
			hlPaint := &core.Paint{Color: color.RGBA{A: 60}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: tabX, Y: 0, Width: tabW, Height: b.Height}, hlPaint)
		} else if tl.hoveredTab == i {
			hlPaint := &core.Paint{Color: color.RGBA{A: 30}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: tabX, Y: 0, Width: tabW, Height: b.Height}, hlPaint)
		}

		// Tab text (centered)
		tabPaint := &core.Paint{Color: textColor, FontSize: fontSize}
		if i == tl.selectedTab {
			tabPaint.FontWeight = 700
		} else {
			tabPaint.Color.A = 180 // slightly dimmed for unselected
		}
		textSize := canvas.MeasureText(tab.Text, tabPaint)
		textX := tabX + (tabW-textSize.Width)/2
		textY := (b.Height - indicatorHeight - textSize.Height) / 2
		canvas.DrawText(tab.Text, textX, textY, tabPaint)
	}

	// 3. Selected tab indicator
	if tl.selectedTab >= 0 && tl.selectedTab < len(tl.tabs) {
		indicatorX := float64(tl.selectedTab) * tabW
		indicatorPaint := &core.Paint{
			Color:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
			DrawStyle: core.PaintFill,
		}
		canvas.DrawRect(core.Rect{
			X:      indicatorX,
			Y:      b.Height - indicatorHeight,
			Width:  tabW,
			Height: indicatorHeight,
		}, indicatorPaint)
	}
}

// ---------- tabLayoutHandler ----------

type tabLayoutHandler struct {
	core.DefaultHandler
	tl *TabLayout
}

func (h *tabLayoutHandler) hitTestTab(node *core.Node, x, y float64) int {
	tl := h.tl
	if len(tl.tabs) == 0 {
		return -1
	}
	// Convert global (window) coordinates to local (node) coordinates
	pos := node.AbsolutePosition()
	localX := x - pos.X
	localY := y - pos.Y
	b := node.Bounds()
	if localY < 0 || localY > b.Height {
		return -1
	}
	tabW := b.Width / float64(len(tl.tabs))
	index := int(localX / tabW)
	if index < 0 || index >= len(tl.tabs) {
		return -1
	}
	return index
}

func (h *tabLayoutHandler) OnEvent(node *core.Node, event core.Event) bool {
	tl := h.tl
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		hit := h.hitTestTab(node, me.X, me.Y)
		if hit >= 0 {
			tl.pressedTab = hit
			node.MarkDirty()
			return true
		}

	case core.ActionUp:
		if tl.pressedTab >= 0 {
			hit := h.hitTestTab(node, me.X, me.Y)
			pressed := tl.pressedTab
			tl.pressedTab = -1
			node.MarkDirty()
			if hit == pressed && node.IsEnabled() {
				tl.SetSelectedTab(hit)
			}
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		hit := h.hitTestTab(node, me.X, me.Y)
		if hit != tl.hoveredTab {
			tl.hoveredTab = hit
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		tl.hoveredTab = -1
		tl.pressedTab = -1
		node.MarkDirty()
		return true
	}

	return false
}
