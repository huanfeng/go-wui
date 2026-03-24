package widget

import (
	"image/color"

	"github.com/huanfeng/go-wui/core"
)

// Divider is a simple horizontal or vertical line separator.
type Divider struct {
	BaseView
}

// NewDivider creates a new Divider with default styling (1dp height, light gray).
func NewDivider() *Divider {
	d := &Divider{}
	d.node = initNode("Divider", d)
	d.node.SetPainter(&dividerPainter{})
	d.node.SetStyle(&core.Style{
		BackgroundColor: core.ParseColor("#E0E0E0"),
		Height:          core.Dimension{Value: 1, Unit: core.DimensionDp},
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
	})
	return d
}

// SetColor sets the divider color.
func (d *Divider) SetColor(c color.RGBA) {
	d.node.GetStyle().BackgroundColor = c
	d.node.MarkDirty()
}

// dividerPainter measures and draws the divider line.
type dividerPainter struct{}

func (p *dividerPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	s := node.GetStyle()

	// Default: horizontal divider — full parent width, 1dp height
	w := 0.0
	h := 1.0

	if s != nil && s.Height.Value > 0 {
		h = s.Height.Value
	}

	// Width: match parent if possible
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	} else if ws.Mode == core.MeasureModeAtMost {
		w = ws.Size
	}

	// Height constraints
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	} else if hs.Mode == core.MeasureModeAtMost && h > hs.Size {
		h = hs.Size
	}

	return core.Size{Width: w, Height: h}
}

func (p *dividerPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	bgColor := s.BackgroundColor
	if bgColor.A == 0 {
		bgColor = core.ParseColor("#E0E0E0")
	}

	paint := &core.Paint{Color: bgColor, DrawStyle: core.PaintFill}
	canvas.DrawRect(localRect, paint)
}
