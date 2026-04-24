package layout

import "github.com/huanfeng/wind-ui/core"

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
			// 读取子节点的 margin，margin 由父布局消耗
			margin := child.Margin()
			marginH := margin.Left + margin.Right
			marginV := margin.Top + margin.Bottom

			var childWidthSpec, childHeightSpec core.MeasureSpec

			if sl.Direction == Vertical {
				// Vertical scroll: child gets parent's width spec, unbound height
				childWidthSpec = childMeasureSpec(widthSpec, paddingH+marginH, dimOrDefault(style, true))
				childHeightSpec = core.MeasureSpec{Mode: core.MeasureModeUnbound}
			} else {
				// Horizontal scroll: child gets unbound width, parent's height spec
				childWidthSpec = core.MeasureSpec{Mode: core.MeasureModeUnbound}
				childHeightSpec = childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
			}

			sz := MeasureChild(child, childWidthSpec, childHeightSpec)
			// 记录子节点含 margin 的整体尺寸，用于滚动范围计算
			sl.childSize = core.Size{
				Width:  sz.Width + marginH,
				Height: sz.Height + marginV,
			}
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
	// 子节点位置 = padding + margin - scroll offset
	margin := child.Margin()
	child.SetBounds(core.Rect{
		X:      padding.Left + margin.Left - sl.OffsetX,
		Y:      padding.Top + margin.Top - sl.OffsetY,
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
