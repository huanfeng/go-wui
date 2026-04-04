package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
)

// Spinner is a dropdown selection widget. It displays the currently selected
// item and opens a PopupMenu overlay when clicked to show all options.
// Modeled after Android's Spinner.
type Spinner struct {
	BaseView
	items       []string
	selectedIdx int
	onSelected  func(index int, item string)

	hovered bool
	pressed bool
}

// NewSpinner creates a new Spinner with the given items.
func NewSpinner(items []string) *Spinner {
	sp := &Spinner{
		items:       items,
		selectedIdx: -1,
	}
	if len(items) > 0 {
		sp.selectedIdx = 0
	}
	sp.node = initNode("Spinner", sp)
	sp.node.SetPainter(&spinnerPainter{sp: sp})
	sp.node.SetHandler(&spinnerHandler{sp: sp})
	sp.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 33, G: 33, B: 33, A: 255},
		FontSize:        14,
		BorderColor:     color.RGBA{R: 180, G: 180, B: 180, A: 255},
		BorderWidth:     1,
		CornerRadius:    4,
	})
	return sp
}

// SetItems replaces the items list.
func (sp *Spinner) SetItems(items []string) {
	sp.items = items
	if sp.selectedIdx >= len(items) {
		sp.selectedIdx = len(items) - 1
	}
	if sp.selectedIdx < 0 && len(items) > 0 {
		sp.selectedIdx = 0
	}
	sp.node.MarkDirty()
}

// GetItems returns the current items list.
func (sp *Spinner) GetItems() []string {
	return sp.items
}

// SetSelectedIndex sets the selected item by index.
func (sp *Spinner) SetSelectedIndex(index int) {
	if index < 0 || index >= len(sp.items) {
		return
	}
	sp.selectedIdx = index
	sp.node.MarkDirty()
}

// GetSelectedIndex returns the index of the selected item.
func (sp *Spinner) GetSelectedIndex() int {
	return sp.selectedIdx
}

// GetSelectedItem returns the currently selected item string.
func (sp *Spinner) GetSelectedItem() string {
	if sp.selectedIdx >= 0 && sp.selectedIdx < len(sp.items) {
		return sp.items[sp.selectedIdx]
	}
	return ""
}

// SetOnItemSelectedListener sets the callback for item selection.
func (sp *Spinner) SetOnItemSelectedListener(fn func(index int, item string)) {
	sp.onSelected = fn
}

// showDropdown opens the PopupMenu overlay with all items.
func (sp *Spinner) showDropdown() {
	menu := NewMenu()
	for i, item := range sp.items {
		idx := i
		text := item
		menu.AddItem("", text, func() {
			sp.selectedIdx = idx
			sp.node.MarkDirty()
			if sp.onSelected != nil {
				sp.onSelected(idx, text)
			}
		})
	}
	pm := NewPopupMenu(menu)
	pos := sp.node.AbsolutePosition()
	b := sp.node.Bounds()
	pm.ShowAtPosition(sp.node, pos.X, pos.Y+b.Height)
}

// ---------- spinnerPainter ----------

type spinnerPainter struct {
	sp *Spinner
}

func (p *spinnerPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	s := node.GetStyle()
	fontSize := 14.0
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize
	}
	w := 150.0
	h := fontSize*1.4 + 16 // text + vertical padding

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

func (p *spinnerPainter) Paint(node *core.Node, canvas core.Canvas) {
	sp := p.sp
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}
	dpi := getDPIScale(node)

	// Background — slightly tinted on hover/press
	bgColor := s.BackgroundColor
	if sp.pressed {
		bgColor = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	} else if sp.hovered {
		bgColor = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	}
	bgPaint := &core.Paint{Color: bgColor, DrawStyle: core.PaintFill}
	canvas.DrawRoundRect(localRect, s.CornerRadius, bgPaint)

	// Border
	if s.BorderWidth > 0 && s.BorderColor.A > 0 {
		borderPaint := &core.Paint{Color: s.BorderColor, DrawStyle: core.PaintStroke, StrokeWidth: s.BorderWidth}
		canvas.DrawRoundRect(localRect, s.CornerRadius, borderPaint)
	}

	fontSize := s.FontSize
	if fontSize == 0 {
		fontSize = 14
	}
	fontSize *= dpi

	// Selected item text
	text := sp.GetSelectedItem()
	if text == "" {
		text = "Select..."
	}
	textPaint := &core.Paint{Color: s.TextColor, FontSize: fontSize}
	textSize := canvas.MeasureText(text, textPaint)
	textY := (b.Height - textSize.Height) / 2
	canvas.DrawText(text, 12*dpi, textY, textPaint)

	// Dropdown arrow "▼" on the right
	arrowPaint := &core.Paint{Color: s.TextColor, FontSize: fontSize * 0.8}
	arrow := "\u25BC"
	arrowSize := canvas.MeasureText(arrow, arrowPaint)
	canvas.DrawText(arrow, b.Width-arrowSize.Width-12*dpi, textY, arrowPaint)
}

// ---------- spinnerHandler ----------

type spinnerHandler struct {
	core.DefaultHandler
	sp *Spinner
}

func (h *spinnerHandler) OnEvent(node *core.Node, event core.Event) bool {
	sp := h.sp
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		sp.pressed = true
		node.MarkDirty()
		return true

	case core.ActionUp:
		wasPressed := sp.pressed
		sp.pressed = false
		node.MarkDirty()
		if wasPressed && node.IsEnabled() && len(sp.items) > 0 {
			sp.showDropdown()
		}
		return true

	case core.ActionHoverEnter:
		sp.hovered = true
		node.MarkDirty()
		return true

	case core.ActionHoverExit, core.ActionCancel:
		sp.hovered = false
		sp.pressed = false
		node.MarkDirty()
		return true
	}
	return false
}
