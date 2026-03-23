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
	li.RegisterView("ScrollView", inflateScrollView)
	li.RegisterView("HorizontalScrollView", inflateHorizontalScrollView)
	li.RegisterView("View", inflateView)
	li.RegisterView("TextView", inflateTextView)
	li.RegisterView("ImageView", inflateImageView)
	li.RegisterView("Button", inflateButton)
	li.RegisterView("Divider", inflateDivider)
	li.RegisterView("CheckBox", inflateCheckBox)
	li.RegisterView("Switch", inflateSwitch)
	li.RegisterView("RadioButton", inflateRadioButton)
	li.RegisterView("RadioGroup", inflateRadioGroup)
	li.RegisterView("ProgressBar", inflateProgressBar)
	li.RegisterView("EditText", inflateEditText)

	// Phase 3 — List & Navigation
	li.RegisterView("Toolbar", inflateToolbar)
	li.RegisterView("TabLayout", inflateTabLayout)
	li.RegisterView("ViewPager", inflateViewPager)
	li.RegisterView("RecyclerView", inflateRecyclerView)
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

	// Apply gravity for text alignment
	switch attrs.GetString("gravity") {
	case "center":
		tv.Node().GetStyle().Gravity = core.GravityCenter
	case "end", "right":
		tv.Node().GetStyle().Gravity = core.GravityEnd
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

func inflateDivider(attrs *AttributeSet) *core.Node {
	d := widget.NewDivider()
	applyCommonAttrs(d.Node(), attrs)
	return d.Node()
}

func inflateCheckBox(attrs *AttributeSet) *core.Node {
	text := attrs.GetString("text")
	cb := widget.NewCheckBox(text)
	applyCommonAttrs(cb.Node(), attrs)

	if attrs.GetBool("checked") {
		cb.SetChecked(true)
	}

	// Apply text-specific attributes
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		cb.Node().GetStyle().FontSize = size.Value
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		cb.Node().GetStyle().TextColor = clr
	}

	return cb.Node()
}

func inflateSwitch(attrs *AttributeSet) *core.Node {
	sw := widget.NewSwitch()
	applyCommonAttrs(sw.Node(), attrs)

	if attrs.GetBool("checked") {
		sw.SetOn(true)
	}

	return sw.Node()
}

func inflateRadioButton(attrs *AttributeSet) *core.Node {
	text := attrs.GetString("text")
	rb := widget.NewRadioButton(text)
	applyCommonAttrs(rb.Node(), attrs)

	if attrs.GetBool("selected") {
		rb.SetSelected(true)
	}

	// Apply text-specific attributes
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		rb.Node().GetStyle().FontSize = size.Value
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		rb.Node().GetStyle().TextColor = clr
	}

	return rb.Node()
}

func inflateRadioGroup(attrs *AttributeSet) *core.Node {
	rg := widget.NewRadioGroup()
	applyCommonAttrs(rg.Node(), attrs)
	return rg.Node()
}

func inflateScrollView(attrs *AttributeSet) *core.Node {
	sv := widget.NewScrollView()
	applyCommonAttrs(sv.Node(), attrs)
	return sv.Node()
}

func inflateHorizontalScrollView(attrs *AttributeSet) *core.Node {
	sv := widget.NewHorizontalScrollView()
	applyCommonAttrs(sv.Node(), attrs)
	return sv.Node()
}

func inflateProgressBar(attrs *AttributeSet) *core.Node {
	pb := widget.NewProgressBar()
	applyCommonAttrs(pb.Node(), attrs)

	if progress := attrs.GetFloat("progress"); progress > 0 {
		pb.SetProgress(progress)
	}
	if attrs.GetBool("indeterminate") {
		pb.SetIndeterminate(true)
	}

	return pb.Node()
}

func inflateEditText(attrs *AttributeSet) *core.Node {
	hint := attrs.GetString("hint")
	et := widget.NewEditText(hint)
	applyCommonAttrs(et.Node(), attrs)

	if text := attrs.GetString("text"); text != "" {
		et.SetText(text)
	}
	// Apply text-specific attributes
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		et.Node().GetStyle().FontSize = size.Value
	}

	return et.Node()
}

func inflateToolbar(attrs *AttributeSet) *core.Node {
	title := attrs.GetString("title")
	tb := widget.NewToolbar(title)
	applyCommonAttrs(tb.Node(), attrs)

	if subtitle := attrs.GetString("subtitle"); subtitle != "" {
		tb.SetSubtitle(subtitle)
	}
	if size := attrs.GetDimension("textSize"); size.Value > 0 {
		tb.Node().GetStyle().FontSize = size.Value
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		tb.Node().GetStyle().TextColor = clr
	}
	if bg := attrs.GetColor("backgroundColor"); bg.A > 0 {
		tb.Node().GetStyle().BackgroundColor = bg
	}

	return tb.Node()
}

func inflateTabLayout(attrs *AttributeSet) *core.Node {
	tl := widget.NewTabLayout()
	applyCommonAttrs(tl.Node(), attrs)

	if bg := attrs.GetColor("backgroundColor"); bg.A > 0 {
		tl.Node().GetStyle().BackgroundColor = bg
	}
	if clr := attrs.GetColor("textColor"); clr.A > 0 {
		tl.Node().GetStyle().TextColor = clr
	}

	return tl.Node()
}

func inflateViewPager(attrs *AttributeSet) *core.Node {
	vp := widget.NewViewPager()
	applyCommonAttrs(vp.Node(), attrs)
	return vp.Node()
}

func inflateRecyclerView(attrs *AttributeSet) *core.Node {
	itemHeight := attrs.GetDimension("itemHeight")
	h := 48.0
	if itemHeight.Value > 0 {
		h = itemHeight.Value
	}
	rv := widget.NewRecyclerView(h)
	applyCommonAttrs(rv.Node(), attrs)
	return rv.Node()
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
