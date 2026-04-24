package layout

import "github.com/huanfeng/wind-ui/core"

// FlexWrap controls whether children wrap to the next line.
type FlexWrap int

const (
	FlexNoWrap FlexWrap = iota
	FlexWrapOn
)

// FlexJustify controls main-axis alignment.
type FlexJustify int

const (
	FlexJustifyStart        FlexJustify = iota
	FlexJustifyCenter
	FlexJustifyEnd
	FlexJustifySpaceBetween
	FlexJustifySpaceAround
)

// FlexAlign controls cross-axis alignment.
type FlexAlign int

const (
	FlexAlignStart   FlexAlign = iota
	FlexAlignCenter
	FlexAlignEnd
	FlexAlignStretch
)

// FlexLayout arranges children using a flexbox-like model.
// Direction is controlled by Orientation (Vertical/Horizontal).
// Supports wrapping, main-axis justification, and cross-axis alignment.
type FlexLayout struct {
	Orientation Orientation
	Wrap        FlexWrap
	Justify     FlexJustify
	AlignItems  FlexAlign
	Spacing     float64 // gap between items on the main axis
	LineSpacing float64 // gap between lines when wrapping
}

// ScaleDPI scales dp-valued fields by the DPI factor.
func (f *FlexLayout) ScaleDPI(scale float64) {
	f.Spacing *= scale
	f.LineSpacing *= scale
}

// flexLine represents one row (horizontal) or column (vertical) of flex items.
type flexLine struct {
	children []*core.Node
	mainSize float64 // total main-axis size of children
	crossMax float64 // max cross-axis size among children
}

// Measure computes the size of the flex container.
func (f *FlexLayout) Measure(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	var visible []*core.Node
	for _, child := range node.Children() {
		if child.GetVisibility() != core.Gone {
			visible = append(visible, child)
		}
	}

	// Measure all children with AtMost specs
	mainAvail := widthSpec.Size - paddingH
	crossAvail := heightSpec.Size - paddingV
	if f.Orientation == Vertical {
		mainAvail = heightSpec.Size - paddingV
		crossAvail = widthSpec.Size - paddingH
	}

	for _, child := range visible {
		// 读取子节点的 margin，margin 由父布局消耗
		margin := child.Margin()
		marginH := margin.Left + margin.Right
		marginV := margin.Top + margin.Bottom

		childWS := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: widthSpec.Size - paddingH - marginH}
		childHS := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: heightSpec.Size - paddingV - marginV}
		style := child.GetStyle()
		if style != nil {
			childWS = childMeasureSpec(widthSpec, paddingH+marginH, dimOrDefault(style, true))
			childHS = childMeasureSpec(heightSpec, paddingV+marginV, dimOrDefault(style, false))
		}
		MeasureChild(child, childWS, childHS)
	}

	// Build lines
	lines := f.buildLines(visible, mainAvail)

	// Compute total size
	totalCross := 0.0
	maxMain := 0.0
	for i, line := range lines {
		if line.mainSize > maxMain {
			maxMain = line.mainSize
		}
		totalCross += line.crossMax
		if i > 0 {
			totalCross += f.LineSpacing
		}
	}

	var resultW, resultH float64
	if f.Orientation == Horizontal {
		resultW = maxMain + paddingH
		resultH = totalCross + paddingV
	} else {
		resultW = totalCross + paddingH
		resultH = maxMain + paddingV
	}

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
	_ = crossAvail // used indirectly via specs
	return size
}

// Arrange positions children according to flex rules.
func (f *FlexLayout) Arrange(node *core.Node, bounds core.Rect) {
	padding := node.Padding()

	var visible []*core.Node
	for _, child := range node.Children() {
		if child.GetVisibility() != core.Gone {
			visible = append(visible, child)
		}
	}

	mainAvail := bounds.Width - padding.Left - padding.Right
	crossAvail := bounds.Height - padding.Top - padding.Bottom
	if f.Orientation == Vertical {
		mainAvail = bounds.Height - padding.Top - padding.Bottom
		crossAvail = bounds.Width - padding.Left - padding.Right
	}

	lines := f.buildLines(visible, mainAvail)

	crossOffset := 0.0
	for _, line := range lines {
		mainOffset := f.justifyOffset(line, mainAvail)
		gap := f.justifyGap(line, mainAvail)

		for _, child := range line.children {
			sz := child.MeasuredSize()
			margin := child.Margin()

			// 计算当前子节点的 margin 在主轴/交叉轴上的分量
			var mainMarginStart, mainMarginEnd, crossMarginStart, crossMarginEnd float64
			if f.Orientation == Horizontal {
				mainMarginStart = margin.Left
				mainMarginEnd = margin.Right
				crossMarginStart = margin.Top
				crossMarginEnd = margin.Bottom
			} else {
				mainMarginStart = margin.Top
				mainMarginEnd = margin.Bottom
				crossMarginStart = margin.Left
				crossMarginEnd = margin.Right
			}

			childMain := f.mainDim(sz)
			childCross := f.crossDim(sz)
			// 子节点在行中占用的交叉轴空间（含 margin）
			childCrossWithMargin := childCross + crossMarginStart + crossMarginEnd

			// Cross-axis alignment：在 crossMax 内以 margin 偏移为基础对齐
			crossPos := crossOffset + crossMarginStart
			switch f.AlignItems {
			case FlexAlignCenter:
				crossPos = crossOffset + crossMarginStart + (line.crossMax-childCrossWithMargin)/2
			case FlexAlignEnd:
				crossPos = crossOffset + line.crossMax - crossMarginEnd - childCross
			case FlexAlignStretch:
				// 拉伸时子节点填满 crossMax 减去两端 margin
				childCross = line.crossMax - crossMarginStart - crossMarginEnd
				if childCross < 0 {
					childCross = 0
				}
				crossPos = crossOffset + crossMarginStart
			}

			var x, y, w, h float64
			if f.Orientation == Horizontal {
				x = padding.Left + mainOffset + mainMarginStart
				y = padding.Top + crossPos
				w = childMain
				h = childCross
			} else {
				x = padding.Left + crossPos
				y = padding.Top + mainOffset + mainMarginStart
				w = childCross
				h = childMain
			}

			child.SetBounds(core.Rect{X: x, Y: y, Width: w, Height: h})
			if l := child.GetLayout(); l != nil {
				l.Arrange(child, child.Bounds())
			}

			// mainOffset 推进时包含子节点主轴尺寸 + 两端主轴 margin + gap
			mainOffset += mainMarginStart + childMain + mainMarginEnd + gap
		}

		crossOffset += line.crossMax + f.LineSpacing
	}
	_ = crossAvail
}

// buildLines groups children into lines based on wrap and available main space.
// margin 参与主轴占用空间和交叉轴最大尺寸的计算。
func (f *FlexLayout) buildLines(children []*core.Node, mainAvail float64) []flexLine {
	if len(children) == 0 {
		return nil
	}

	var lines []flexLine
	var current flexLine

	for _, child := range children {
		sz := child.MeasuredSize()
		margin := child.Margin()
		// 子节点在主轴上占用的总空间 = 测量尺寸 + 两端 margin
		var childMainMargin, childCrossMargin float64
		if f.Orientation == Horizontal {
			childMainMargin = margin.Left + margin.Right
			childCrossMargin = margin.Top + margin.Bottom
		} else {
			childMainMargin = margin.Top + margin.Bottom
			childCrossMargin = margin.Left + margin.Right
		}
		childMain := f.mainDim(sz) + childMainMargin
		childCross := f.crossDim(sz) + childCrossMargin

		// Check if wrapping is needed
		if f.Wrap == FlexWrapOn && len(current.children) > 0 {
			spaceNeeded := current.mainSize + f.Spacing + childMain
			if spaceNeeded > mainAvail {
				lines = append(lines, current)
				current = flexLine{}
			}
		}

		if len(current.children) > 0 {
			current.mainSize += f.Spacing
		}
		current.mainSize += childMain
		if childCross > current.crossMax {
			current.crossMax = childCross
		}
		current.children = append(current.children, child)
	}

	if len(current.children) > 0 {
		lines = append(lines, current)
	}
	return lines
}

// justifyOffset returns the starting offset for items in a line.
func (f *FlexLayout) justifyOffset(line flexLine, mainAvail float64) float64 {
	freeSpace := mainAvail - line.mainSize
	if freeSpace < 0 {
		freeSpace = 0
	}
	switch f.Justify {
	case FlexJustifyCenter:
		return freeSpace / 2
	case FlexJustifyEnd:
		return freeSpace
	case FlexJustifySpaceAround:
		if len(line.children) > 0 {
			return freeSpace / float64(len(line.children)) / 2
		}
	}
	return 0
}

// justifyGap returns the gap between items based on justify mode.
func (f *FlexLayout) justifyGap(line flexLine, mainAvail float64) float64 {
	freeSpace := mainAvail - line.mainSize
	if freeSpace < 0 {
		freeSpace = 0
	}
	n := len(line.children)
	switch f.Justify {
	case FlexJustifySpaceBetween:
		if n > 1 {
			return freeSpace/float64(n-1) + f.Spacing
		}
	case FlexJustifySpaceAround:
		if n > 0 {
			return freeSpace/float64(n) + f.Spacing
		}
	}
	return f.Spacing
}

// mainDim returns the main-axis size component.
func (f *FlexLayout) mainDim(sz core.Size) float64 {
	if f.Orientation == Horizontal {
		return sz.Width
	}
	return sz.Height
}

// crossDim returns the cross-axis size component.
func (f *FlexLayout) crossDim(sz core.Size) float64 {
	if f.Orientation == Horizontal {
		return sz.Height
	}
	return sz.Width
}
