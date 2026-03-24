package widget

import "github.com/huanfeng/go-wui/core"

// ViewWidget is a basic view that draws a background and border.
type ViewWidget struct {
	BaseView
}

// NewView creates a new basic ViewWidget.
func NewView() *ViewWidget {
	v := &ViewWidget{}
	v.node = initNode("View", v)
	v.node.SetPainter(&viewPainter{})
	v.node.SetStyle(&core.Style{})
	return v
}

// viewPainter draws background color, border, and corner radius.
type viewPainter struct{}

func (p *viewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Pure container — return 0x0, let layout handle sizing
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *viewPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	// Draw background
	if s.BackgroundColor.A > 0 {
		paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}
	// Draw border
	if s.BorderWidth > 0 && s.BorderColor.A > 0 {
		paint := &core.Paint{Color: s.BorderColor, DrawStyle: core.PaintStroke, StrokeWidth: s.BorderWidth}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}
}
