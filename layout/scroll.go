package layout

import "github.com/huanfeng/go-wui/core"

// ScrollLayout measures its single child with an unbound constraint on the
// scroll axis, allowing the child to extend beyond the viewport. It stores
// the scroll offset which is used during arrange and paint.
type ScrollLayout struct {
	Direction Orientation // Vertical or Horizontal
	OffsetX   float64     // current horizontal scroll offset
	OffsetY   float64     // current vertical scroll offset

	// childSize stores the child's measured size for scroll clamping.
	childSize core.Size
}

// Measure computes the desired size of the scroll container. The single child
// is measured with an Unbound spec on the scroll axis so it can be larger than
// the viewport.
func (sl *ScrollLayout) Measure(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	children := node.Children()
	if len(children) > 0 {
		child := children[0]
		if child.GetVisibility() == core.Gone {
			// No visible child — just use viewport size
		} else {
			style := child.GetStyle()

			var childWidthSpec, childHeightSpec core.MeasureSpec

			if sl.Direction == Vertical {
				// Vertical scroll: child gets parent's width spec, unbound height
				childWidthSpec = childMeasureSpec(widthSpec, paddingH, dimOrDefault(style, true))
				childHeightSpec = core.MeasureSpec{Mode: core.MeasureModeUnbound}
			} else {
				// Horizontal scroll: child gets unbound width, parent's height spec
				childWidthSpec = core.MeasureSpec{Mode: core.MeasureModeUnbound}
				childHeightSpec = childMeasureSpec(heightSpec, paddingV, dimOrDefault(style, false))
			}

			sz := MeasureChild(child, childWidthSpec, childHeightSpec)
			sl.childSize = sz
		}
	}

	// The ScrollLayout itself takes the viewport size from the parent spec.
	var resultW, resultH float64

	switch widthSpec.Mode {
	case core.MeasureModeExact:
		resultW = widthSpec.Size
	case core.MeasureModeAtMost:
		resultW = widthSpec.Size
		if sl.childSize.Width+paddingH < resultW {
			resultW = sl.childSize.Width + paddingH
		}
	default:
		resultW = sl.childSize.Width + paddingH
	}

	switch heightSpec.Mode {
	case core.MeasureModeExact:
		resultH = heightSpec.Size
	case core.MeasureModeAtMost:
		resultH = heightSpec.Size
		if sl.childSize.Height+paddingV < resultH {
			resultH = sl.childSize.Height + paddingV
		}
	default:
		resultH = sl.childSize.Height + paddingV
	}

	size := core.Size{Width: resultW, Height: resultH}
	node.SetMeasuredSize(size)
	return size
}

// Arrange positions the single child offset by the current scroll position.
func (sl *ScrollLayout) Arrange(node *core.Node, bounds core.Rect) {
	padding := node.Padding()

	children := node.Children()
	if len(children) == 0 {
		return
	}
	child := children[0]
	if child.GetVisibility() == core.Gone {
		return
	}

	sz := child.MeasuredSize()
	child.SetBounds(core.Rect{
		X:      padding.Left - sl.OffsetX,
		Y:      padding.Top - sl.OffsetY,
		Width:  sz.Width,
		Height: sz.Height,
	})

	// Recursively arrange children that have their own layout
	if l := child.GetLayout(); l != nil {
		l.Arrange(child, child.Bounds())
	}
}

// ChildSize returns the measured size of the scroll content child.
func (sl *ScrollLayout) ChildSize() core.Size {
	return sl.childSize
}
