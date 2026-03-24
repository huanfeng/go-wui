package widget

import (
	"image/color"

	"github.com/huanfeng/go-wui/core"
)

// MenuItem represents a single item in a menu.
type MenuItem struct {
	ID      string
	Title   string
	Enabled bool
	OnClick func()
}

// Menu is a list of MenuItems. It serves as the data model for PopupMenu.
type Menu struct {
	items []MenuItem
}

// NewMenu creates an empty Menu.
func NewMenu() *Menu {
	return &Menu{}
}

// Add appends a menu item.
func (m *Menu) Add(item MenuItem) *Menu {
	if item.Enabled == false && item.OnClick != nil {
		// Default enabled to true if handler set but enabled wasn't explicit
	}
	m.items = append(m.items, item)
	return m
}

// AddItem is a convenience method to add a simple menu item.
func (m *Menu) AddItem(id, title string, onClick func()) *Menu {
	m.items = append(m.items, MenuItem{
		ID:      id,
		Title:   title,
		Enabled: true,
		OnClick: onClick,
	})
	return m
}

// GetItems returns all menu items.
func (m *Menu) GetItems() []MenuItem {
	return m.items
}

// GetItemCount returns the number of items.
func (m *Menu) GetItemCount() int {
	return len(m.items)
}

// Clear removes all items.
func (m *Menu) Clear() {
	m.items = nil
}

// ---------- PopupMenu ----------

const (
	menuItemHeight  = 40.0
	menuItemPadding = 16.0
	menuMinWidth    = 150.0
	menuMaxWidth    = 300.0
	menuCornerRadius = 4.0
)

// PopupMenu shows a floating menu anchored to a position.
// It renders as an overlay node added to the root of the node tree.
type PopupMenu struct {
	BaseView
	menu        *Menu
	anchorX     float64
	anchorY     float64
	onDismiss   func()
	showing     bool
	overlayNode *core.Node // the full-screen overlay

	hoveredItem int
	pressedItem int
}

// NewPopupMenu creates a PopupMenu with the given menu data.
func NewPopupMenu(menu *Menu) *PopupMenu {
	pm := &PopupMenu{
		menu:        menu,
		hoveredItem: -1,
		pressedItem: -1,
	}
	pm.node = initNode("PopupMenu", pm)
	pm.node.SetPainter(&popupMenuPainter{pm: pm})
	pm.node.SetHandler(&popupMenuHandler{pm: pm})
	pm.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 33, G: 33, B: 33, A: 255},
		FontSize:        14,
	})
	return pm
}

// ShowAtPosition shows the popup menu at the given absolute coordinates.
// It creates an overlay node and adds it to the root of the anchor's node tree.
func (pm *PopupMenu) ShowAtPosition(anchor *core.Node, x, y float64) {
	if pm.showing {
		return
	}
	pm.anchorX = x
	pm.anchorY = y
	pm.showing = true

	// Find root node
	root := anchor
	for root.Parent() != nil {
		root = root.Parent()
	}

	// Create full-screen overlay — marked so layout systems skip it and PaintNode renders it on top
	pm.overlayNode = core.NewNode("PopupMenuOverlay")
	pm.overlayNode.SetPainter(&popupOverlayPainter{pm: pm})
	pm.overlayNode.SetHandler(&popupOverlayHandler{pm: pm})
	pm.overlayNode.SetData("paintsChildren", true)
	pm.overlayNode.SetData("isOverlay", true)

	// Add menu node as child of overlay
	pm.overlayNode.AddChild(pm.node)

	// Set overlay bounds immediately so event dispatch can hit-test it
	rootSize := root.MeasuredSize()
	rootBounds := root.Bounds()
	w := rootSize.Width
	if w <= 0 {
		w = rootBounds.Width
	}
	h := rootSize.Height
	if h <= 0 {
		h = rootBounds.Height
	}
	pm.overlayNode.SetBounds(core.Rect{X: 0, Y: 0, Width: w, Height: h})
	pm.overlayNode.SetMeasuredSize(core.Size{Width: w, Height: h})

	// Add overlay to root
	root.AddChild(pm.overlayNode)
}

// Dismiss closes the popup menu and removes the overlay.
func (pm *PopupMenu) Dismiss() {
	if !pm.showing {
		return
	}
	pm.showing = false
	if pm.overlayNode != nil && pm.overlayNode.Parent() != nil {
		pm.overlayNode.Parent().RemoveChild(pm.overlayNode)
	}
	pm.overlayNode = nil
	if pm.onDismiss != nil {
		pm.onDismiss()
	}
}

// IsShowing returns whether the menu is currently displayed.
func (pm *PopupMenu) IsShowing() bool {
	return pm.showing
}

// SetOnDismissListener sets the callback invoked when the menu is dismissed.
func (pm *PopupMenu) SetOnDismissListener(fn func()) {
	pm.onDismiss = fn
}

// menuWidth calculates the width based on the longest item title (DPI-scaled).
func (pm *PopupMenu) menuWidth() float64 {
	dpi := getDPIScale(pm.node)
	s := pm.node.GetStyle()
	fontSize := 14.0 * dpi // fallback
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize // already DPI-scaled
	}
	charWidth := fontSize * 0.6
	maxW := menuMinWidth * dpi
	for _, item := range pm.menu.items {
		w := float64(len([]rune(item.Title)))*charWidth + menuItemPadding*2*dpi
		if w > maxW {
			maxW = w
		}
	}
	maxMenuW := menuMaxWidth * dpi
	if maxW > maxMenuW {
		maxW = maxMenuW
	}
	return maxW
}

// menuHeight returns the total height of all menu items (DPI-scaled).
func (pm *PopupMenu) menuHeight() float64 {
	dpi := getDPIScale(pm.node)
	return float64(len(pm.menu.items)) * menuItemHeight * dpi
}

// ---------- popupOverlayPainter (full-screen backdrop) ----------

type popupOverlayPainter struct {
	pm *PopupMenu
}

func (p *popupOverlayPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *popupOverlayPainter) Paint(node *core.Node, canvas core.Canvas) {
	pm := p.pm
	b := node.Bounds()

	// Semi-transparent backdrop (very light)
	backdropPaint := &core.Paint{Color: color.RGBA{A: 20}, DrawStyle: core.PaintFill}
	canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, backdropPaint)

	// Position and paint the menu
	menuW := pm.menuWidth()
	menuH := pm.menuHeight()

	// Clamp to window bounds
	menuX := pm.anchorX
	menuY := pm.anchorY
	if menuX+menuW > b.Width {
		menuX = b.Width - menuW
	}
	if menuY+menuH > b.Height {
		menuY = b.Height - menuH
	}
	if menuX < 0 {
		menuX = 0
	}
	if menuY < 0 {
		menuY = 0
	}

	pm.node.SetBounds(core.Rect{X: menuX, Y: menuY, Width: menuW, Height: menuH})
	paintNodeRecursive(pm.node, canvas)
}

// ---------- popupOverlayHandler (captures outside clicks) ----------

type popupOverlayHandler struct {
	core.DefaultHandler
	pm *PopupMenu
}

func (h *popupOverlayHandler) OnEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return true // consume non-motion events (modal)
	}
	if me.Action == core.ActionDown {
		// Check if click is outside menu bounds using absolute coordinates
		mb := h.pm.node.Bounds()
		pos := h.pm.node.AbsolutePosition()
		absRect := core.Rect{X: pos.X, Y: pos.Y, Width: mb.Width, Height: mb.Height}
		if !absRect.Contains(me.X, me.Y) {
			h.pm.Dismiss()
			return true
		}
	}
	return true // consume all events (modal)
}

// ---------- popupMenuPainter ----------

type popupMenuPainter struct {
	pm *PopupMenu
}

func (p *popupMenuPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	return core.Size{Width: p.pm.menuWidth(), Height: p.pm.menuHeight()}
}

func (p *popupMenuPainter) Paint(node *core.Node, canvas core.Canvas) {
	pm := p.pm
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	dpi := getDPIScale(node)
	cr := menuCornerRadius * dpi
	itemH := menuItemHeight * dpi
	pad := menuItemPadding * dpi

	// Menu background with shadow effect (slightly darker border)
	bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
	canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, cr, bgPaint)

	borderPaint := &core.Paint{
		Color:       color.RGBA{R: 200, G: 200, B: 200, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 1,
	}
	canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, cr, borderPaint)

	// Draw items
	fontSize := 14.0 * dpi // fallback
	if s.FontSize > 0 {
		fontSize = s.FontSize // already DPI-scaled
	}
	for i, item := range pm.menu.items {
		itemY := float64(i) * itemH
		itemRect := core.Rect{X: 0, Y: itemY, Width: b.Width, Height: itemH}

		// Hover/press highlight
		if pm.pressedItem == i {
			hlPaint := &core.Paint{Color: color.RGBA{R: 0, G: 0, B: 0, A: 30}, DrawStyle: core.PaintFill}
			canvas.DrawRect(itemRect, hlPaint)
		} else if pm.hoveredItem == i {
			hlPaint := &core.Paint{Color: color.RGBA{R: 0, G: 0, B: 0, A: 15}, DrawStyle: core.PaintFill}
			canvas.DrawRect(itemRect, hlPaint)
		}

		// Item text
		textColor := s.TextColor
		if !item.Enabled {
			textColor.A = 100 // dimmed
		}
		textPaint := &core.Paint{Color: textColor, FontSize: fontSize}
		textSize := canvas.MeasureText(item.Title, textPaint)
		textY := itemY + (itemH-textSize.Height)/2
		canvas.DrawText(item.Title, pad, textY, textPaint)
	}
}

// ---------- popupMenuHandler ----------

type popupMenuHandler struct {
	core.DefaultHandler
	pm *PopupMenu
}

func (h *popupMenuHandler) hitTestItem(node *core.Node, x, y float64) int {
	pm := h.pm
	if len(pm.menu.items) == 0 {
		return -1
	}
	// Convert global coordinates to local (relative to menu node)
	pos := pm.node.AbsolutePosition()
	localX := x - pos.X
	localY := y - pos.Y
	b := pm.node.Bounds()
	if localX < 0 || localX > b.Width || localY < 0 || localY > b.Height {
		return -1
	}
	dpi := getDPIScale(pm.node)
	index := int(localY / (menuItemHeight * dpi))
	if index < 0 || index >= len(pm.menu.items) {
		return -1
	}
	return index
}

func (h *popupMenuHandler) OnEvent(node *core.Node, event core.Event) bool {
	pm := h.pm
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		hit := h.hitTestItem(node, me.X, me.Y)
		if hit >= 0 && pm.menu.items[hit].Enabled {
			pm.pressedItem = hit
			node.MarkDirty()
			return true
		}

	case core.ActionUp:
		if pm.pressedItem >= 0 {
			hit := h.hitTestItem(node, me.X, me.Y)
			pressed := pm.pressedItem
			pm.pressedItem = -1
			node.MarkDirty()
			if hit == pressed && pm.menu.items[hit].Enabled {
				if pm.menu.items[hit].OnClick != nil {
					pm.menu.items[hit].OnClick()
				}
				pm.Dismiss()
			}
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		hit := h.hitTestItem(node, me.X, me.Y)
		if hit != pm.hoveredItem {
			pm.hoveredItem = hit
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		pm.hoveredItem = -1
		pm.pressedItem = -1
		node.MarkDirty()
		return true
	}

	return false
}
