package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
)

// DialogButton identifies which button was clicked.
type DialogButton int

const (
	DialogButtonPositive DialogButton = iota
	DialogButtonNegative
	DialogButtonNeutral
)

const (
	dialogMaxWidth     = 320.0
	dialogPadding      = 24.0
	dialogButtonHeight = 40.0
	dialogButtonGap    = 8.0
	dialogCornerRadius = 8.0
)

// Dialog is a modal overlay that shows a title, message, and up to three buttons.
// Modeled after Android's AlertDialog.
type Dialog struct {
	BaseView
	title          string
	message        string
	positiveText   string
	negativeText   string
	neutralText    string
	positiveClick  func()
	negativeClick  func()
	neutralClick   func()
	onDismiss      func()
	showing        bool
	overlayNode    *core.Node
	cancelable     bool

	hoveredButton int // -1 = none, 0=positive, 1=negative, 2=neutral
	pressedButton int
}

// AlertDialogBuilder provides a builder pattern for creating Dialog instances.
type AlertDialogBuilder struct {
	dialog *Dialog
}

// NewAlertDialogBuilder creates a new builder for an AlertDialog.
func NewAlertDialogBuilder() *AlertDialogBuilder {
	d := &Dialog{
		hoveredButton: -1,
		pressedButton: -1,
		cancelable:    true,
	}
	d.node = initNode("Dialog", d)
	d.node.SetPainter(&dialogPainter{d: d})
	d.node.SetHandler(&dialogHandler{d: d})
	d.node.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 33, G: 33, B: 33, A: 255},
		FontSize:        14,
	})
	return &AlertDialogBuilder{dialog: d}
}

// SetTitle sets the dialog title.
func (b *AlertDialogBuilder) SetTitle(title string) *AlertDialogBuilder {
	b.dialog.title = title
	return b
}

// SetMessage sets the dialog message.
func (b *AlertDialogBuilder) SetMessage(message string) *AlertDialogBuilder {
	b.dialog.message = message
	return b
}

// SetPositiveButton sets the positive button text and handler.
func (b *AlertDialogBuilder) SetPositiveButton(text string, onClick func()) *AlertDialogBuilder {
	b.dialog.positiveText = text
	b.dialog.positiveClick = onClick
	return b
}

// SetNegativeButton sets the negative button text and handler.
func (b *AlertDialogBuilder) SetNegativeButton(text string, onClick func()) *AlertDialogBuilder {
	b.dialog.negativeText = text
	b.dialog.negativeClick = onClick
	return b
}

// SetNeutralButton sets the neutral button text and handler.
func (b *AlertDialogBuilder) SetNeutralButton(text string, onClick func()) *AlertDialogBuilder {
	b.dialog.neutralText = text
	b.dialog.neutralClick = onClick
	return b
}

// SetCancelable sets whether the dialog can be dismissed by clicking outside.
func (b *AlertDialogBuilder) SetCancelable(cancelable bool) *AlertDialogBuilder {
	b.dialog.cancelable = cancelable
	return b
}

// SetOnDismissListener sets the callback invoked when the dialog is dismissed.
func (b *AlertDialogBuilder) SetOnDismissListener(fn func()) *AlertDialogBuilder {
	b.dialog.onDismiss = fn
	return b
}

// Build returns the configured Dialog without showing it.
func (b *AlertDialogBuilder) Build() *Dialog {
	return b.dialog
}

// Show builds and shows the dialog immediately, attached to the given root node.
func (b *AlertDialogBuilder) Show(root *core.Node) *Dialog {
	d := b.dialog
	d.ShowInNode(root)
	return d
}

// ---------- Dialog methods ----------

// GetTitle returns the dialog title.
func (d *Dialog) GetTitle() string {
	return d.title
}

// GetMessage returns the dialog message.
func (d *Dialog) GetMessage() string {
	return d.message
}

// ShowInNode shows the dialog as an overlay added to the given root node.
func (d *Dialog) ShowInNode(root *core.Node) {
	if d.showing {
		return
	}
	d.showing = true

	// Navigate to the actual root
	for root.Parent() != nil {
		root = root.Parent()
	}

	// Create overlay — marked so layout systems skip it and PaintNode renders it on top
	d.overlayNode = core.NewNode("DialogOverlay")
	d.overlayNode.SetPainter(&dialogOverlayPainter{d: d})
	d.overlayNode.SetHandler(&dialogOverlayHandler{d: d})
	d.overlayNode.SetData("paintsChildren", true)
	d.overlayNode.SetData("isOverlay", true)
	d.overlayNode.AddChild(d.node)

	// Set overlay bounds immediately so event dispatch can hit-test it
	// before the next paint cycle. PaintNode will refresh these during render.
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
	d.overlayNode.SetBounds(core.Rect{X: 0, Y: 0, Width: w, Height: h})
	d.overlayNode.SetMeasuredSize(core.Size{Width: w, Height: h})

	root.AddChild(d.overlayNode)
}

// Dismiss closes the dialog and removes the overlay.
func (d *Dialog) Dismiss() {
	if !d.showing {
		return
	}
	d.showing = false
	if d.overlayNode != nil && d.overlayNode.Parent() != nil {
		d.overlayNode.Parent().RemoveChild(d.overlayNode)
	}
	d.overlayNode = nil
	if d.onDismiss != nil {
		d.onDismiss()
	}
}

// IsShowing returns whether the dialog is currently displayed.
func (d *Dialog) IsShowing() bool {
	return d.showing
}

// buttonTexts returns the button labels in order: neutral, negative, positive (right-aligned).
func (d *Dialog) buttonTexts() []struct {
	text    string
	btnType DialogButton
} {
	var btns []struct {
		text    string
		btnType DialogButton
	}
	if d.neutralText != "" {
		btns = append(btns, struct {
			text    string
			btnType DialogButton
		}{d.neutralText, DialogButtonNeutral})
	}
	if d.negativeText != "" {
		btns = append(btns, struct {
			text    string
			btnType DialogButton
		}{d.negativeText, DialogButtonNegative})
	}
	if d.positiveText != "" {
		btns = append(btns, struct {
			text    string
			btnType DialogButton
		}{d.positiveText, DialogButtonPositive})
	}
	return btns
}

// ---------- dialogOverlayPainter ----------

type dialogOverlayPainter struct {
	d *Dialog
}

func (p *dialogOverlayPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *dialogOverlayPainter) Paint(node *core.Node, canvas core.Canvas) {
	b := node.Bounds()

	// Semi-transparent dark backdrop
	backdropPaint := &core.Paint{Color: color.RGBA{A: 100}, DrawStyle: core.PaintFill}
	canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, backdropPaint)

	// DPI-scaled sizes
	dpi := getDPIScale(node)
	scaledMaxWidth := dialogMaxWidth * dpi
	scaledPadding := dialogPadding * dpi
	scaledButtonHeight := dialogButtonHeight * dpi
	// Calculate dialog size
	d := p.d
	dialogW := scaledMaxWidth
	if dialogW > b.Width-48*dpi {
		dialogW = b.Width - 48*dpi
	}

	s := d.node.GetStyle()
	fontSize := 14.0 * dpi // fallback (constant, needs manual scaling)
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize // already DPI-scaled by AddChild
	}
	titleFontSize := fontSize * 1.3

	// Estimate dialog height
	dialogH := scaledPadding // top padding
	if d.title != "" {
		dialogH += titleFontSize*1.4 + 8*dpi
	}
	if d.message != "" {
		msgPaint := &core.Paint{FontSize: fontSize}
		contentW := dialogW - scaledPadding*2
		msgSize := core.NodeMeasureText(node, d.message, msgPaint)
		// Estimate wrapped lines based on measured text width
		lines := int(msgSize.Width/contentW) + 1
		if lines < 1 {
			lines = 1
		}
		dialogH += float64(lines) * msgSize.Height
	}
	btns := d.buttonTexts()
	if len(btns) > 0 {
		dialogH += 16*dpi + scaledButtonHeight // gap + button row
	}
	dialogH += scaledPadding // bottom padding

	// Center dialog
	dialogX := (b.Width - dialogW) / 2
	dialogY := (b.Height - dialogH) / 2
	if dialogY < 48*dpi {
		dialogY = 48 * dpi
	}

	// Store DPI scale for the dialog painter to use
	d.node.SetData("dpiScale", dpi)
	d.node.SetBounds(core.Rect{X: dialogX, Y: dialogY, Width: dialogW, Height: dialogH})
	d.node.SetMeasuredSize(core.Size{Width: dialogW, Height: dialogH})
	paintNodeRecursive(d.node, canvas)
}

// ---------- dialogOverlayHandler ----------

type dialogOverlayHandler struct {
	core.DefaultHandler
	d *Dialog
}

func (h *dialogOverlayHandler) OnEvent(node *core.Node, event core.Event) bool {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return true // consume all non-motion events to block input to background
	}
	if me.Action == core.ActionDown {
		// Check using absolute coordinates
		mb := h.d.node.Bounds()
		pos := h.d.node.AbsolutePosition()
		absRect := core.Rect{X: pos.X, Y: pos.Y, Width: mb.Width, Height: mb.Height}
		if !absRect.Contains(me.X, me.Y) && h.d.cancelable {
			h.d.Dismiss()
			return true
		}
	}
	return true // consume all events (modal)
}

// ---------- dialogPainter ----------

type dialogPainter struct {
	d *Dialog
}

func (p *dialogPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Size is managed by overlay painter
	return core.Size{Width: dialogMaxWidth, Height: 200}
}

func (p *dialogPainter) Paint(node *core.Node, canvas core.Canvas) {
	d := p.d
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	dpi := getDPIScale(node)
	pad := dialogPadding * dpi
	btnH := dialogButtonHeight * dpi
	btnGap := dialogButtonGap * dpi
	cr := dialogCornerRadius * dpi

	// Dialog background
	bgPaint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
	canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, cr, bgPaint)

	// Border for depth
	borderPaint := &core.Paint{
		Color:       color.RGBA{R: 180, G: 180, B: 180, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 1,
	}
	canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, cr, borderPaint)

	fontSize := 14.0 * dpi // fallback
	if s.FontSize > 0 {
		fontSize = s.FontSize // already DPI-scaled
	}
	titleFontSize := fontSize * 1.3
	textColor := s.TextColor

	y := pad

	// Title
	if d.title != "" {
		titlePaint := &core.Paint{Color: textColor, FontSize: titleFontSize, FontWeight: 700}
		canvas.DrawText(d.title, pad, y, titlePaint)
		y += titleFontSize*1.4 + 8*dpi
	}

	// Message
	if d.message != "" {
		msgPaint := &core.Paint{Color: textColor, FontSize: fontSize}
		msgPaint.Color.A = 200

		// Word wrapping using measured text width
		contentW := b.Width - pad*2
		runes := []rune(d.message)
		lineStart := 0
		for lineStart < len(runes) {
			// Binary-search for the longest substring that fits contentW
			lineEnd := len(runes)
			for lo, hi := lineStart+1, len(runes); lo <= hi; {
				mid := (lo + hi) / 2
				seg := string(runes[lineStart:mid])
				segSize := canvas.MeasureText(seg, msgPaint)
				if segSize.Width <= contentW {
					lineEnd = mid
					lo = mid + 1
				} else {
					hi = mid - 1
				}
			}
			// If the full remaining text fits, take it all
			fullSeg := string(runes[lineStart:])
			if canvas.MeasureText(fullSeg, msgPaint).Width <= contentW {
				lineEnd = len(runes)
			}
			line := string(runes[lineStart:lineEnd])
			lineSize := canvas.MeasureText(line, msgPaint)
			canvas.DrawText(line, pad, y, msgPaint)
			y += lineSize.Height
			lineStart = lineEnd
		}
	}

	// Buttons (right-aligned)
	btns := d.buttonTexts()
	if len(btns) > 0 {
		y = b.Height - pad - btnH
		btnX := b.Width - pad
		primaryColor := core.ParseColor("#1976D2")

		for i := len(btns) - 1; i >= 0; i-- {
			btn := btns[i]
			btnPaint := &core.Paint{Color: primaryColor, FontSize: fontSize, FontWeight: 600}
			btnSize := canvas.MeasureText(btn.text, btnPaint)
			btnW := btnSize.Width + 24*dpi
			btnX -= btnW + btnGap

			btnRect := core.Rect{X: btnX, Y: y, Width: btnW, Height: btnH}

			// Hover/press highlight — use light background so text stays visible
			if d.pressedButton == i {
				hlPaint := &core.Paint{Color: color.RGBA{R: 200, G: 220, B: 245, A: 255}, DrawStyle: core.PaintFill}
				canvas.DrawRoundRect(btnRect, 4*dpi, hlPaint)
			} else if d.hoveredButton == i {
				hlPaint := &core.Paint{Color: color.RGBA{R: 225, G: 235, B: 250, A: 255}, DrawStyle: core.PaintFill}
				canvas.DrawRoundRect(btnRect, 4*dpi, hlPaint)
			}

			textY := y + (btnH-btnSize.Height)/2
			canvas.DrawText(btn.text, btnX+12*dpi, textY, btnPaint)
		}
	}
}

// ---------- dialogHandler ----------

type dialogHandler struct {
	core.DefaultHandler
	d *Dialog
}

func (h *dialogHandler) hitTestButton(node *core.Node, x, y float64) int {
	d := h.d
	b := d.node.Bounds()
	s := d.node.GetStyle()
	dpi := getDPIScale(d.node)
	fontSize := 14.0 * dpi // fallback
	if s != nil && s.FontSize > 0 {
		fontSize = s.FontSize // already DPI-scaled
	}

	// Convert global coordinates to local (relative to dialog node)
	pos := d.node.AbsolutePosition()
	localX := x - pos.X
	localY := y - pos.Y

	pad := dialogPadding * dpi
	btnH := dialogButtonHeight * dpi
	btnGap := dialogButtonGap * dpi

	btns := d.buttonTexts()
	if len(btns) == 0 {
		return -1
	}

	btnY := b.Height - pad - btnH
	if localY < btnY || localY > btnY+btnH {
		return -1
	}

	btnX := b.Width - pad
	btnPaint := &core.Paint{FontSize: fontSize, FontWeight: 600}
	for i := len(btns) - 1; i >= 0; i-- {
		btn := btns[i]
		btnTextSize := core.NodeMeasureText(d.node, btn.text, btnPaint)
		btnW := btnTextSize.Width + 24*dpi
		btnX -= btnW + btnGap
		if localX >= btnX && localX < btnX+btnW {
			return i
		}
	}
	return -1
}

func (h *dialogHandler) OnEvent(node *core.Node, event core.Event) bool {
	d := h.d
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return true
	}

	switch me.Action {
	case core.ActionDown:
		hit := h.hitTestButton(node, me.X, me.Y)
		if hit >= 0 {
			d.pressedButton = hit
			node.MarkDirty()
		}
		return true

	case core.ActionUp:
		if d.pressedButton >= 0 {
			hit := h.hitTestButton(node, me.X, me.Y)
			pressed := d.pressedButton
			d.pressedButton = -1
			node.MarkDirty()
			if hit == pressed {
				btns := d.buttonTexts()
				if pressed < len(btns) {
					switch btns[pressed].btnType {
					case DialogButtonPositive:
						if d.positiveClick != nil {
							d.positiveClick()
						}
					case DialogButtonNegative:
						if d.negativeClick != nil {
							d.negativeClick()
						}
					case DialogButtonNeutral:
						if d.neutralClick != nil {
							d.neutralClick()
						}
					}
					d.Dismiss()
				}
			}
		}
		return true

	case core.ActionHoverEnter, core.ActionHoverMove:
		hit := h.hitTestButton(node, me.X, me.Y)
		if hit != d.hoveredButton {
			d.hoveredButton = hit
			node.MarkDirty()
		}
		return true

	case core.ActionHoverExit, core.ActionCancel:
		d.hoveredButton = -1
		d.pressedButton = -1
		node.MarkDirty()
		return true
	}

	return true
}
