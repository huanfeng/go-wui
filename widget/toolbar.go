package widget

import (
	"image/color"

	"gowui/core"
)

// defaultToolbarHeight is the standard toolbar height in dp (Android convention).
const defaultToolbarHeight = 56.0

// ActionItem represents a single action button in the toolbar.
type ActionItem struct {
	ID      string
	Title   string
	OnClick func()
}

// Toolbar is a horizontal bar that displays a title, optional subtitle,
// optional navigation button (left), and action items (right).
// Modeled after Android's Toolbar.
type Toolbar struct {
	BaseView
	title      string
	subtitle   string
	navOnClick func()
	navText    string // navigation button text (e.g. "←")
	actions    []ActionItem

	// hover/press tracking
	hoveredAction int // -1 = none, -2 = nav
	pressedAction int // -1 = none, -2 = nav
}

// NewToolbar creates a new Toolbar with the given title.
func NewToolbar(title string) *Toolbar {
	tb := &Toolbar{
		title:         title,
		hoveredAction: -1,
		pressedAction: -1,
	}
	tb.node = initNode("Toolbar", tb)
	tb.node.SetPainter(&toolbarPainter{tb: tb})
	tb.node.SetHandler(&toolbarHandler{tb: tb})
	tb.node.SetStyle(&core.Style{
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		FontSize:        20,
	})
	return tb
}

// SetTitle sets the toolbar title.
func (tb *Toolbar) SetTitle(title string) {
	tb.title = title
	tb.node.MarkDirty()
}

// GetTitle returns the toolbar title.
func (tb *Toolbar) GetTitle() string {
	return tb.title
}

// SetSubtitle sets the toolbar subtitle.
func (tb *Toolbar) SetSubtitle(subtitle string) {
	tb.subtitle = subtitle
	tb.node.MarkDirty()
}

// GetSubtitle returns the toolbar subtitle.
func (tb *Toolbar) GetSubtitle() string {
	return tb.subtitle
}

// SetNavigationOnClickListener sets the click handler for the navigation button.
// Pass a non-nil handler to show the navigation button.
func (tb *Toolbar) SetNavigationOnClickListener(fn func()) {
	tb.navOnClick = fn
	if tb.navText == "" {
		tb.navText = "\u2190" // ← arrow
	}
	tb.node.MarkDirty()
}

// SetNavigationText sets the text displayed on the navigation button.
func (tb *Toolbar) SetNavigationText(text string) {
	tb.navText = text
	tb.node.MarkDirty()
}

// AddAction adds an action item to the right side of the toolbar.
func (tb *Toolbar) AddAction(item ActionItem) {
	tb.actions = append(tb.actions, item)
	tb.node.MarkDirty()
}

// ClearActions removes all action items.
func (tb *Toolbar) ClearActions() {
	tb.actions = nil
	tb.node.MarkDirty()
}

// GetActionCount returns the number of action items.
func (tb *Toolbar) GetActionCount() int {
	return len(tb.actions)
}

// ---------- toolbarPainter ----------

type toolbarPainter struct {
	tb *Toolbar
}

func (p *toolbarPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	h := defaultToolbarHeight
	w := 300.0 // default minimum width

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

func (p *toolbarPainter) Paint(node *core.Node, canvas core.Canvas) {
	tb := p.tb
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// 1. Background
	bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
	canvas.DrawRect(localRect, bgPaint)

	titleFontSize := s.FontSize
	if titleFontSize == 0 {
		titleFontSize = 20
	}
	subtitleFontSize := titleFontSize * 0.7
	textColor := s.TextColor
	if textColor.A == 0 {
		textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	xOffset := 16.0 // left padding

	// 2. Navigation button
	if tb.navOnClick != nil && tb.navText != "" {
		navPaint := &core.Paint{Color: textColor, FontSize: titleFontSize}
		navSize := canvas.MeasureText(tb.navText, navPaint)
		navWidth := navSize.Width + 16 // padding around nav text

		// Highlight on hover/press
		if tb.pressedAction == -2 {
			hlPaint := &core.Paint{Color: color.RGBA{A: 60}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: xOffset - 8, Y: 0, Width: navWidth + 8, Height: b.Height}, hlPaint)
		} else if tb.hoveredAction == -2 {
			hlPaint := &core.Paint{Color: color.RGBA{A: 30}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: xOffset - 8, Y: 0, Width: navWidth + 8, Height: b.Height}, hlPaint)
		}

		navY := (b.Height - navSize.Height) / 2
		canvas.DrawText(tb.navText, xOffset, navY, navPaint)
		xOffset += navWidth + 8
	}

	// 3. Title and subtitle
	if tb.title != "" {
		titlePaint := &core.Paint{Color: textColor, FontSize: titleFontSize, FontWeight: 700}
		if tb.subtitle != "" {
			// Two-line: title above, subtitle below
			titleSize := canvas.MeasureText(tb.title, titlePaint)
			subtitlePaint := &core.Paint{Color: textColor, FontSize: subtitleFontSize}
			subtitlePaint.Color.A = 200 // slightly transparent
			subtitleSize := canvas.MeasureText(tb.subtitle, subtitlePaint)

			totalH := titleSize.Height + subtitleSize.Height + 2
			titleY := (b.Height - totalH) / 2
			canvas.DrawText(tb.title, xOffset, titleY, titlePaint)
			canvas.DrawText(tb.subtitle, xOffset, titleY+titleSize.Height+2, subtitlePaint)
		} else {
			// Single-line title centered vertically
			titleSize := canvas.MeasureText(tb.title, titlePaint)
			titleY := (b.Height - titleSize.Height) / 2
			canvas.DrawText(tb.title, xOffset, titleY, titlePaint)
		}
	}

	// 4. Action items (right-aligned)
	actionX := b.Width - 16.0 // right padding
	for i := len(tb.actions) - 1; i >= 0; i-- {
		action := tb.actions[i]
		actionPaint := &core.Paint{Color: textColor, FontSize: titleFontSize * 0.8}
		actionSize := canvas.MeasureText(action.Title, actionPaint)
		actionWidth := actionSize.Width + 16
		actionX -= actionWidth

		// Highlight on hover/press
		if tb.pressedAction == i {
			hlPaint := &core.Paint{Color: color.RGBA{A: 60}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: actionX, Y: 0, Width: actionWidth, Height: b.Height}, hlPaint)
		} else if tb.hoveredAction == i {
			hlPaint := &core.Paint{Color: color.RGBA{A: 30}, DrawStyle: core.PaintFill}
			canvas.DrawRect(core.Rect{X: actionX, Y: 0, Width: actionWidth, Height: b.Height}, hlPaint)
		}

		actionY := (b.Height - actionSize.Height) / 2
		canvas.DrawText(action.Title, actionX+8, actionY, actionPaint)
	}
}

// ---------- toolbarHandler ----------

type toolbarHandler struct {
	core.DefaultHandler
	tb *Toolbar
}

// hitTest determines which action item (or nav button) the pointer is over.
// Returns: -2 = nav, 0..n = action index, -1 = nothing.
func (h *toolbarHandler) hitTest(node *core.Node, x, y float64) int {
	tb := h.tb
	b := node.Bounds()
	s := node.GetStyle()
	titleFontSize := 20.0
	if s != nil && s.FontSize > 0 {
		titleFontSize = s.FontSize
	}

	// Convert global (window) coordinates to local (node) coordinates
	pos := node.AbsolutePosition()
	localX := x - pos.X
	localY := y - pos.Y

	xOffset := 16.0

	// Check nav button
	if tb.navOnClick != nil && tb.navText != "" {
		charWidth := titleFontSize * 0.6
		navWidth := float64(len([]rune(tb.navText)))*charWidth + 24
		if localX >= xOffset-8 && localX < xOffset+navWidth && localY >= 0 && localY <= b.Height {
			return -2
		}
		xOffset += navWidth + 8
	}

	// Check action items (right-aligned)
	actionX := b.Width - 16.0
	actionFontSize := titleFontSize * 0.8
	for i := len(tb.actions) - 1; i >= 0; i-- {
		action := tb.actions[i]
		charWidth := actionFontSize * 0.6
		actionWidth := float64(len([]rune(action.Title)))*charWidth + 16
		actionX -= actionWidth
		if localX >= actionX && localX < actionX+actionWidth && localY >= 0 && localY <= b.Height {
			return i
		}
	}

	return -1
}

func (h *toolbarHandler) OnEvent(node *core.Node, event core.Event) bool {
	tb := h.tb
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false
	}

	switch me.Action {
	case core.ActionDown:
		hit := h.hitTest(node, me.X, me.Y)
		if hit != -1 {
			tb.pressedAction = hit
			node.MarkDirty()
			return true
		}

	case core.ActionUp:
		if tb.pressedAction != -1 {
			hit := h.hitTest(node, me.X, me.Y)
			pressed := tb.pressedAction
			tb.pressedAction = -1
			node.MarkDirty()
			if hit == pressed && node.IsEnabled() {
				if hit == -2 && tb.navOnClick != nil {
					tb.navOnClick()
				} else if hit >= 0 && hit < len(tb.actions) && tb.actions[hit].OnClick != nil {
					tb.actions[hit].OnClick()
				}
			}
			return true
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		hit := h.hitTest(node, me.X, me.Y)
		if hit != tb.hoveredAction {
			tb.hoveredAction = hit
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		tb.hoveredAction = -1
		tb.pressedAction = -1
		node.MarkDirty()
		return true
	}

	return false
}
