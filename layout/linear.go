package layout

import "github.com/huanfeng/wind-ui/core"

// Orientation defines the direction of a LinearLayout.
type Orientation int

const (
	Vertical   Orientation = iota // Children stacked top-to-bottom
	Horizontal                    // Children placed left-to-right
)

// LinearLayout arranges children sequentially along a single axis.
// It supports spacing between children, cross-axis gravity alignment,
// and weight-based distribution of remaining space.
type LinearLayout struct {
	Orientation Orientation
	Spacing     float64
	Gravity     core.Gravity // cross-axis alignment for children
}

// ScaleDPI scales dp-valued fields by the DPI factor.
func (ll *LinearLayout) ScaleDPI(scale float64) {
	ll.Spacing *= scale
}

// Measure computes the desired size of the node and all its visible children.
func (ll *LinearLayout) Measure(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	if ll.Orientation == Horizontal {
		return ll.measureHorizontal(node, widthSpec, heightSpec)
	}
	return ll.measureVertical(node, widthSpec, heightSpec)
}

// Arrange positions each visible child within the given bounds.
func (ll *LinearLayout) Arrange(node *core.Node, bounds core.Rect) {
	if ll.Orientation == Horizontal {
		ll.arrangeHorizontal(node, bounds)
	} else {
		ll.arrangeVertical(node, bounds)
	}
}

// ---------- Vertical ----------

func (ll *LinearLayout) measureVertical(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	// Available space for children after padding
	availH := heightSpec.Size - paddingV

	var totalHeight float64
	var maxWidth float64
	var totalWeight float64
	visibleCount := 0

	children := node.Children()

	// First pass: measure non-weighted children, accumulate weights
	for _, child := range children {
		if child.GetVisibility() == core.Gone {
			continue
		}
		// Skip overlay nodes — they are rendered separately on top of the tree
		if child.GetData("isOverlay") != nil {
			continue
		}
		visibleCount++

		// 读取子节点的 margin，margin 由父布局消耗
		margin := child.Margin()
		marginH := margin.Left + margin.Right
		marginV := margin.Top + margin.Bottom

		style := child.GetStyle()
		if style != nil && style.Height.Unit == core.DimensionWeight {
			totalWeight += style.Weight
			// 加权子节点第一遍仅测量宽度，需减去 margin 占用的空间
			childWidthSpec := childMeasureSpec(widthSpec, paddingH+marginH, style.Width)
			MeasureChild(child, childWidthSpec, core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: availH - marginV})
			sz := child.MeasuredSize()
			// 宽度累计需加上子节点的水平 margin
			if sz.Width+marginH > maxWidth {
				maxWidth = sz.Width + marginH
			}
			continue
		}

		childWidthSpec := childMeasureSpec(widthSpec, paddingH+marginH, dimOrDefault(style, true))
		childHeightSpec := childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
		sz := MeasureChild(child, childWidthSpec, childHeightSpec)

		// 子节点在主轴上占用的空间 = 测量高度 + 垂直 margin
		totalHeight += sz.Height + marginV
		// 子节点在交叉轴上占用的空间 = 测量宽度 + 水平 margin
		if sz.Width+marginH > maxWidth {
			maxWidth = sz.Width + marginH
		}
	}

	// Add spacing between visible children
	if visibleCount > 1 {
		totalHeight += ll.Spacing * float64(visibleCount-1)
	}

	// Second pass: distribute remaining space to weighted children
	if totalWeight > 0 && heightSpec.Mode == core.MeasureModeExact {
		remaining := availH - totalHeight
		if remaining < 0 {
			remaining = 0
		}
		for _, child := range children {
			if child.GetVisibility() == core.Gone {
				continue
			}
			style := child.GetStyle()
			if style == nil || style.Height.Unit != core.DimensionWeight {
				continue
			}
			margin := child.Margin()
			marginH := margin.Left + margin.Right
			marginV := margin.Top + margin.Bottom
			// 加权子节点分得的空间需减去其自身 margin
			portion := remaining * style.Weight / totalWeight
			childWidthSpec := childMeasureSpec(widthSpec, paddingH+marginH, style.Width)
			sz := MeasureChild(child, childWidthSpec, core.MeasureSpec{Mode: core.MeasureModeExact, Size: portion - marginV})
			if sz.Width+marginH > maxWidth {
				maxWidth = sz.Width + marginH
			}
			totalHeight += portion
		}
	}

	// Resolve final size
	resultW := maxWidth + paddingH
	resultH := totalHeight + paddingV

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

func (ll *LinearLayout) arrangeVertical(node *core.Node, bounds core.Rect) {
	padding := node.Padding()
	// Child bounds are RELATIVE to parent — start from padding, not bounds.X/Y
	contentX := padding.Left
	contentW := bounds.Width - padding.Left - padding.Right
	curY := padding.Top

	for _, child := range node.Children() {
		if child.GetVisibility() == core.Gone {
			continue
		}
		if child.GetData("isOverlay") != nil {
			continue
		}
		sz := child.MeasuredSize()
		margin := child.Margin()

		// 主轴：先推进 margin.Top，子节点从 curY+margin.Top 开始放置
		childY := curY + margin.Top

		// 交叉轴（水平）对齐，在 padding 内容区中以 margin 缩减后的宽度计算偏移
		innerW := contentW - margin.Left - margin.Right
		x := contentX + margin.Left
		switch ll.Gravity {
		case core.GravityCenter:
			x = contentX + margin.Left + (innerW-sz.Width)/2
		case core.GravityEnd:
			x = contentX + margin.Left + innerW - sz.Width
		}

		child.SetBounds(core.Rect{
			X:      x,
			Y:      childY,
			Width:  sz.Width,
			Height: sz.Height,
		})

		// Recursively arrange children that have their own layout
		if l := child.GetLayout(); l != nil {
			l.Arrange(child, child.Bounds())
		}

		// 推进 curY：子节点高度 + margin.Top + margin.Bottom + Spacing
		curY += margin.Top + sz.Height + margin.Bottom + ll.Spacing
	}
}

// ---------- Horizontal ----------

func (ll *LinearLayout) measureHorizontal(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	availW := widthSpec.Size - paddingH

	var totalWidth float64
	var maxHeight float64
	var totalWeight float64
	visibleCount := 0

	children := node.Children()

	// First pass: measure non-weighted children, accumulate weights
	for _, child := range children {
		if child.GetVisibility() == core.Gone {
			continue
		}
		if child.GetData("isOverlay") != nil {
			continue
		}
		visibleCount++

		// 读取子节点的 margin，margin 由父布局消耗
		margin := child.Margin()
		marginH := margin.Left + margin.Right
		marginV := margin.Top + margin.Bottom

		style := child.GetStyle()
		if style != nil && style.Width.Unit == core.DimensionWeight {
			totalWeight += style.Weight
			// 加权子节点第一遍仅测量高度，需减去 margin 占用的空间
			childHeightSpec := childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
			MeasureChild(child, core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: availW - marginH}, childHeightSpec)
			sz := child.MeasuredSize()
			// 高度累计需加上子节点的垂直 margin
			if sz.Height+marginV > maxHeight {
				maxHeight = sz.Height + marginV
			}
			continue
		}

		childWidthSpec := childMeasureSpec(widthSpec, paddingH+marginH, dimOrDefault(style, true))
		childHeightSpec := childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
		sz := MeasureChild(child, childWidthSpec, childHeightSpec)

		// 子节点在主轴上占用的空间 = 测量宽度 + 水平 margin
		totalWidth += sz.Width + marginH
		// 子节点在交叉轴上占用的空间 = 测量高度 + 垂直 margin
		if sz.Height+marginV > maxHeight {
			maxHeight = sz.Height + marginV
		}
	}

	// Add spacing between visible children
	if visibleCount > 1 {
		totalWidth += ll.Spacing * float64(visibleCount-1)
	}

	// Second pass: distribute remaining space to weighted children
	if totalWeight > 0 && widthSpec.Mode == core.MeasureModeExact {
		remaining := availW - totalWidth
		if remaining < 0 {
			remaining = 0
		}
		for _, child := range children {
			if child.GetVisibility() == core.Gone {
				continue
			}
			style := child.GetStyle()
			if style == nil || style.Width.Unit != core.DimensionWeight {
				continue
			}
			margin := child.Margin()
			marginH := margin.Left + margin.Right
			marginV := margin.Top + margin.Bottom
			// 加权子节点分得的空间需减去其自身 margin
			portion := remaining * style.Weight / totalWeight
			childHeightSpec := childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
			sz := MeasureChild(child, core.MeasureSpec{Mode: core.MeasureModeExact, Size: portion - marginH}, childHeightSpec)
			if sz.Height+marginV > maxHeight {
				maxHeight = sz.Height + marginV
			}
			totalWidth += portion
		}
	}

	// Resolve final size
	resultW := totalWidth + paddingH
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

func (ll *LinearLayout) arrangeHorizontal(node *core.Node, bounds core.Rect) {
	padding := node.Padding()
	// Child bounds are RELATIVE to parent — start from padding, not bounds.X/Y
	contentY := padding.Top
	contentH := bounds.Height - padding.Top - padding.Bottom
	curX := padding.Left

	for _, child := range node.Children() {
		if child.GetVisibility() == core.Gone {
			continue
		}
		if child.GetData("isOverlay") != nil {
			continue
		}
		sz := child.MeasuredSize()
		margin := child.Margin()

		// 主轴：先推进 margin.Left，子节点从 curX+margin.Left 开始放置
		childX := curX + margin.Left

		// 交叉轴（垂直）对齐，在 padding 内容区中以 margin 缩减后的高度计算偏移
		innerH := contentH - margin.Top - margin.Bottom
		y := contentY + margin.Top
		switch ll.Gravity {
		case core.GravityCenter:
			y = contentY + margin.Top + (innerH-sz.Height)/2
		case core.GravityEnd:
			y = contentY + margin.Top + innerH - sz.Height
		}

		child.SetBounds(core.Rect{
			X:      childX,
			Y:      y,
			Width:  sz.Width,
			Height: sz.Height,
		})

		// Recursively arrange children that have their own layout
		if l := child.GetLayout(); l != nil {
			l.Arrange(child, child.Bounds())
		}

		// 推进 curX：子节点宽度 + margin.Left + margin.Right + Spacing
		curX += margin.Left + sz.Width + margin.Right + ll.Spacing
	}
}

// ---------- Helpers ----------

// MeasureChild measures a single child node using its layout or painter.
// It sets the child's MeasuredSize and returns the resulting size.
func MeasureChild(child *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	if l := child.GetLayout(); l != nil {
		size := l.Measure(child, widthSpec, heightSpec)
		child.SetMeasuredSize(size)
		return size
	}
	if p := child.GetPainter(); p != nil {
		size := p.Measure(child, widthSpec, heightSpec)
		child.SetMeasuredSize(size)
		return size
	}
	return core.Size{} // no layout, no painter = 0x0
}

// childMeasureSpec determines the MeasureSpec for a child along one axis
// based on the parent's spec and the child's Dimension setting.
func childMeasureSpec(parentSpec core.MeasureSpec, usedSpace float64, dim core.Dimension) core.MeasureSpec {
	available := parentSpec.Size - usedSpace
	if available < 0 {
		available = 0
	}

	switch dim.Unit {
	case core.DimensionMatchParent:
		switch parentSpec.Mode {
		case core.MeasureModeExact:
			return core.MeasureSpec{Mode: core.MeasureModeExact, Size: available}
		case core.MeasureModeAtMost:
			return core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: available}
		default:
			return core.MeasureSpec{Mode: core.MeasureModeUnbound}
		}

	case core.DimensionDp, core.DimensionPx:
		return core.MeasureSpec{Mode: core.MeasureModeExact, Size: dim.Value}

	case core.DimensionWrapContent:
		switch parentSpec.Mode {
		case core.MeasureModeExact, core.MeasureModeAtMost:
			return core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: available}
		default:
			return core.MeasureSpec{Mode: core.MeasureModeUnbound}
		}

	default: // includes DimensionWeight — handled separately
		switch parentSpec.Mode {
		case core.MeasureModeExact, core.MeasureModeAtMost:
			return core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: available}
		default:
			return core.MeasureSpec{Mode: core.MeasureModeUnbound}
		}
	}
}

// dimOrDefault returns the appropriate Dimension from a child's style.
// isWidth selects Width or Height.
// A zero-value Dimension (DimensionPx with Value 0) is treated as WrapContent
// since it indicates the dimension was never explicitly set.
func dimOrDefault(style *core.Style, isWidth bool) core.Dimension {
	if style == nil {
		return core.Dimension{Unit: core.DimensionWrapContent}
	}
	var d core.Dimension
	if isWidth {
		d = style.Width
	} else {
		d = style.Height
	}
	// Treat zero-value Dimension as WrapContent (not explicitly set)
	if d == (core.Dimension{}) {
		return core.Dimension{Unit: core.DimensionWrapContent}
	}
	return d
}
