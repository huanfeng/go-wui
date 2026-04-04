package main

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
)

// ClipItemWidget is a custom View that draws a Win11-style clipboard card.
// Used as RecyclerView item — paints card background, text preview, source info,
// and pin indicator all in one Painter.
type ClipItemWidget struct {
	node *core.Node
}

func (w *ClipItemWidget) Node() *core.Node               { return w.node }
func (w *ClipItemWidget) SetId(id string)                 { w.node.SetId(id) }
func (w *ClipItemWidget) GetId() string                   { return w.node.GetId() }
func (w *ClipItemWidget) SetVisibility(v core.Visibility) { w.node.SetVisibility(v) }
func (w *ClipItemWidget) GetVisibility() core.Visibility  { return w.node.GetVisibility() }
func (w *ClipItemWidget) SetEnabled(e bool)               { w.node.SetEnabled(e) }
func (w *ClipItemWidget) IsEnabled() bool                 { return w.node.IsEnabled() }

// NewClipItemWidget creates a new clipboard card widget.
func NewClipItemWidget() *ClipItemWidget {
	w := &ClipItemWidget{}
	w.node = core.NewNode("ClipItem")
	w.node.SetView(w)
	w.node.SetPainter(&clipItemPainter{})
	return w
}

// SetContent updates the display content.
func (w *ClipItemWidget) SetContent(text, source, timeStr string, pinned bool) {
	w.node.SetData("itemText", text)
	w.node.SetData("itemSource", source)
	w.node.SetData("itemTime", timeStr)
	if pinned {
		w.node.SetData("itemPinned", true)
	} else {
		w.node.SetData("itemPinned", nil)
	}
	w.node.MarkDirty()
}

// clipItemPainter draws a Win11-style clipboard card.
type clipItemPainter struct{}

func (p *clipItemPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w := 300.0
	h := 64.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *clipItemPainter) Paint(node *core.Node, canvas core.Canvas) {
	b := node.Bounds()

	// All sizes are proportional to the item bounds (already DPI-scaled by framework).
	// Use ratios instead of manual DPI multiplication.
	h := b.Height
	marginH := h * 0.1
	marginV := h * 0.04
	radius := h * 0.12
	padX := h * 0.2
	padY := h * 0.14

	// Card area
	cardRect := core.Rect{
		X:      marginH,
		Y:      marginV,
		Width:  b.Width - marginH*2,
		Height: h - marginV*2,
	}

	// Card background (white)
	canvas.DrawRoundRect(cardRect, radius, &core.Paint{
		Color:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		DrawStyle: core.PaintFill,
	})

	// Card border
	canvas.DrawRoundRect(cardRect, radius, &core.Paint{
		Color:       color.RGBA{R: 230, G: 230, B: 230, A: 255},
		DrawStyle:   core.PaintStroke,
		StrokeWidth: 1,
	})

	text := node.GetDataString("itemText")
	if text == "" {
		return
	}

	// Font sizes proportional to item height (already DPI-scaled).
	fontSize := h * 0.24
	smallFontSize := h * 0.17

	// Main text
	textPaint := &core.Paint{
		Color:    color.RGBA{R: 26, G: 26, B: 26, A: 255},
		FontSize: fontSize,
	}
	textX := cardRect.X + padX
	textY := cardRect.Y + padY
	canvas.DrawText(text, textX, textY, textPaint)

	// Source + time
	source := node.GetDataString("itemSource")
	timeStr := node.GetDataString("itemTime")
	if source != "" || timeStr != "" {
		secText := source
		if timeStr != "" {
			if secText != "" {
				secText += " · "
			}
			secText += timeStr
		}
		secPaint := &core.Paint{
			Color:    color.RGBA{R: 150, G: 150, B: 150, A: 255},
			FontSize: smallFontSize,
		}
		canvas.DrawText(secText, textX, textY+fontSize*1.5, secPaint)
	}

	// Pin indicator
	if node.GetData("itemPinned") != nil {
		pinPaint := &core.Paint{
			Color:    color.RGBA{R: 0, G: 120, B: 212, A: 255},
			FontSize: fontSize,
		}
		pinX := cardRect.X + cardRect.Width - padX - fontSize
		pinY := cardRect.Y + cardRect.Height/2 - fontSize/2
		canvas.DrawText("★", pinX, pinY, pinPaint)
	}
}

func getDPIScaleFromNode(node *core.Node) float64 {
	for n := node; n != nil; n = n.Parent() {
		if s, ok := n.GetData("dpiScale").(float64); ok && s > 0 {
			return s
		}
	}
	return 1.0
}
