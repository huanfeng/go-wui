package layout

import "github.com/huanfeng/wind-ui/core"

// FrameLayout stacks all children on top of each other.
// Each child's position is controlled by its Style.Gravity.
type FrameLayout struct{}

// Measure computes the desired size as the maximum of all visible children.
func (fl *FrameLayout) Measure(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	var maxWidth, maxHeight float64

	for _, child := range node.Children() {
		if child.GetVisibility() == core.Gone {
			continue
		}

		// 读取子节点的 margin，margin 由父布局消耗
		margin := child.Margin()
		marginH := margin.Left + margin.Right
		marginV := margin.Top + margin.Bottom

		style := child.GetStyle()
		childWidthSpec := childMeasureSpec(widthSpec, paddingH+marginH, dimOrDefault(style, true))
		childHeightSpec := childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))

		sz := MeasureChild(child, childWidthSpec, childHeightSpec)
		// 子节点在容器中占用的空间包含其自身 margin
		if sz.Width+marginH > maxWidth {
			maxWidth = sz.Width + marginH
		}
		if sz.Height+marginV > maxHeight {
			maxHeight = sz.Height + marginV
		}
	}

	// Resolve final size
	resultW := maxWidth + paddingH
	resultH := maxHeight + paddingV

	switch widthSpec.Mode {
	case core.MeasureModeExact:
		resultW = widthSpec.Size
	case core.MeasureModeAtMost:
		if resultW > widthSpec.Size {
			resultW = widthSpec.Size
		}
	}

	switch heightSpec.Mode {
	case core.MeasureModeExact:
		resultH = heightSpec.Size
	case core.MeasureModeAtMost:
		if resultH > heightSpec.Size {
			resultH = heightSpec.Size
		}
	}

	size := core.Size{Width: resultW, Height: resultH}
	node.SetMeasuredSize(size)
	return size
}

// Arrange positions each visible child based on its Style.Gravity.
func (fl *FrameLayout) Arrange(node *core.Node, bounds core.Rect) {
	padding := node.Padding()
	// Child bounds are RELATIVE to parent — start from padding, not bounds.X/Y
	contentX := padding.Left
	contentY := padding.Top
	contentW := bounds.Width - padding.Left - padding.Right
	contentH := bounds.Height - padding.Top - padding.Bottom

	for _, child := range node.Children() {
		if child.GetVisibility() == core.Gone {
			continue
		}

		sz := child.MeasuredSize()
		style := child.GetStyle()
		// 读取子节点的 margin，margin 由父布局消耗
		margin := child.Margin()

		gravity := core.GravityStart
		if style != nil {
			gravity = style.Gravity
		}

		// 内容区减去 margin 后的可用空间
		innerW := contentW - margin.Left - margin.Right
		innerH := contentH - margin.Top - margin.Bottom

		var x, y float64

		switch gravity {
		case core.GravityCenter:
			x = contentX + margin.Left + (innerW-sz.Width)/2
			y = contentY + margin.Top + (innerH-sz.Height)/2
		case core.GravityEnd:
			x = contentX + margin.Left + innerW - sz.Width
			y = contentY + margin.Top + innerH - sz.Height
		default: // GravityStart
			x = contentX + margin.Left
			y = contentY + margin.Top
		}

		child.SetBounds(core.Rect{
			X:      x,
			Y:      y,
			Width:  sz.Width,
			Height: sz.Height,
		})

		// Recursively arrange children that have their own layout
		if l := child.GetLayout(); l != nil {
			l.Arrange(child, child.Bounds())
		}
	}
}
