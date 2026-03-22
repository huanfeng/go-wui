package res

import (
	"gowui/core"
	"gowui/layout"
	"gowui/widget"
)

// RegisterBuiltinViews registers all built-in view factories with the inflater.
func RegisterBuiltinViews(li *LayoutInflater) {
	li.RegisterView("LinearLayout", inflateLinearLayout)
	li.RegisterView("FrameLayout", inflateFrameLayout)
	li.RegisterView("View", inflateView)
	li.RegisterView("TextView", inflateTextView)
	li.RegisterView("ImageView", inflateImageView)
	li.RegisterView("Button", inflateButton)
}

func inflateLinearLayout(attrs *AttributeSet) *core.Node {
	n := core.NewNode("LinearLayout")
	n.SetStyle(&core.Style{})

	orientation := layout.Vertical
	if attrs.GetString("orientation") == "horizontal" {
		orientation = layout.Horizontal
	}

	spacing := attrs.GetDimension("spacing")

	gravity := core.GravityStart
	switch attrs.GetString("gravity") {
	case "center":
		gravity = core.GravityCenter
	case "end", "right", "bottom":
		gravity = core.GravityEnd
	case "center_vertical":
		gravity = core.GravityCenterVertical
	case "center_horizontal":
		gravity = core.GravityCenterHorizontal
	}

	ll := &layout.LinearLayout{
		Orientation: orientation,
		Spacing:     spacing.Value,
		Gravity:     gravity,
	}
	n.SetLayout(ll)
	n.SetPainter(&containerPainter{})
	applyCommonAttrs(n, attrs)
	return n
}

func inflateFrameLayout(attrs *AttributeSet) *core.Node {
	n := core.NewNode("FrameLayout")
	n.SetStyle(&core.Style{})

	fl := &layout.FrameLayout{}
	n.SetLayout(fl)
	n.SetPainter(&containerPainter{})
	applyCommonAttrs(n, attrs)
	return n
}

func inflateView(attrs *AttributeSet) *core.Node {
	v := widget.NewView()
	applyCommonAttrs(v.Node(), attrs)
	return v.Node()
}

func inflateTextView(attrs *AttributeSet) *core.Node {
	text := attrs.GetString("text")
	tv := widget.NewTextView(text)
	applyCommonAttrs(tv.Node(), attrs)

	// Apply text-specific attributes
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		tv.Node().GetStyle().FontSize = size.Value
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		tv.Node().GetStyle().TextColor = clr
	}

	return tv.Node()
}

func inflateImageView(attrs *AttributeSet) *core.Node {
	iv := widget.NewImageView()
	applyCommonAttrs(iv.Node(), attrs)
	return iv.Node()
}

func inflateButton(attrs *AttributeSet) *core.Node {
	text := attrs.GetString("text")
	btn := widget.NewButton(text, nil)
	applyCommonAttrs(btn.Node(), attrs)

	// Apply button-specific attributes
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		btn.Node().GetStyle().FontSize = size.Value
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		btn.Node().GetStyle().TextColor = clr
	}

	return btn.Node()
}

// containerPainter is a simple painter for layout containers that draws
// background color. Used by LinearLayout and FrameLayout inflated from XML.
type containerPainter struct{}

func (p *containerPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Container delegates measurement to its Layout
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *containerPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}

	if s.BackgroundColor.A > 0 {
		paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}
}
