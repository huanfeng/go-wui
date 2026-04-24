package layout

import "github.com/huanfeng/wind-ui/core"

// GridLayout arranges children in a rows × columns grid.
// Children are placed left-to-right, top-to-bottom.
// If ColumnCount is set, rows are computed automatically.
// Supports uniform cell sizing and spacing.
type GridLayout struct {
	ColumnCount int     // number of columns (required, >= 1)
	Spacing     float64 // gap between cells (both horizontal and vertical)
}

// ScaleDPI scales dp-valued fields by the DPI factor.
func (g *GridLayout) ScaleDPI(scale float64) {
	g.Spacing *= scale
}

// Measure computes the grid size based on columns and number of children.
func (g *GridLayout) Measure(node *core.Node, widthSpec, heightSpec core.MeasureSpec) core.Size {
	cols := g.ColumnCount
	if cols < 1 {
		cols = 1
	}
	padding := node.Padding()
	paddingH := padding.Left + padding.Right
	paddingV := padding.Top + padding.Bottom

	// Count visible children
	var visible []*core.Node
	for _, child := range node.Children() {
		if child.GetVisibility() != core.Gone {
			visible = append(visible, child)
		}
	}

	rows := (len(visible) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}

	// Calculate cell size
	availW := widthSpec.Size - paddingH
	if availW < 0 {
		availW = 0
	}
	totalSpacingH := g.Spacing * float64(cols-1)
	cellW := (availW - totalSpacingH) / float64(cols)
	if cellW < 0 {
		cellW = 0
	}

	// Measure each child with cell-sized width and style-aware height
	// 注意：GridLayout 使用统一单元格尺寸，margin 在单元格内部消耗
	maxCellH := 0.0
	childWS := core.MeasureSpec{Mode: core.MeasureModeExact, Size: cellW}
	for _, child := range visible {
		margin := child.Margin()
		marginV := margin.Top + margin.Bottom
		marginH := margin.Left + margin.Right
		// 单元格内子节点的可用宽度需减去水平 margin
		effectiveCellW := cellW - marginH
		if effectiveCellW < 0 {
			effectiveCellW = 0
		}
		effectiveWS := core.MeasureSpec{Mode: core.MeasureModeExact, Size: effectiveCellW}
		// Respect child's style Height (e.g. "48dp") via childMeasureSpec
		childHS := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: heightSpec.Size}
		style := child.GetStyle()
		if style != nil {
			childHS = childMeasureSpec(heightSpec, marginV, dimOrDefault(style, false))
		}
		sz := MeasureChild(child, effectiveWS, childHS)
		// 单元格高度包含子节点高度及其垂直 margin
		if sz.Height+marginV > maxCellH {
			maxCellH = sz.Height + marginV
		}
	}

	// All cells have uniform height (tallest cell wins)
	// Re-measure with exact height (子节点高度 = 单元格高度 - 垂直 margin)
	for _, child := range visible {
		margin := child.Margin()
		marginV := margin.Top + margin.Bottom
		marginH := margin.Left + margin.Right
		effectiveCellW := cellW - marginH
		if effectiveCellW < 0 {
			effectiveCellW = 0
		}
		effectiveWS := core.MeasureSpec{Mode: core.MeasureModeExact, Size: effectiveCellW}
		childH := maxCellH - marginV
		if childH < 0 {
			childH = 0
		}
		childHS := core.MeasureSpec{Mode: core.MeasureModeExact, Size: childH}
		MeasureChild(child, effectiveWS, childHS)
	}
	_ = childWS // 保留变量以避免编译错误

	totalSpacingV := g.Spacing * float64(rows-1)
	resultW := availW + paddingH
	resultH := float64(rows)*maxCellH + totalSpacingV + paddingV

	if widthSpec.Mode == core.MeasureModeExact {
		resultW = widthSpec.Size
	}
	if heightSpec.Mode == core.MeasureModeExact {
		resultH = heightSpec.Size
	} else if heightSpec.Mode == core.MeasureModeAtMost && resultH > heightSpec.Size {
		resultH = heightSpec.Size
	}

	size := core.Size{Width: resultW, Height: resultH}
	node.SetMeasuredSize(size)
	return size
}

// Arrange positions children in the grid.
func (g *GridLayout) Arrange(node *core.Node, bounds core.Rect) {
	cols := g.ColumnCount
	if cols < 1 {
		cols = 1
	}
	padding := node.Padding()
	availW := bounds.Width - padding.Left - padding.Right
	totalSpacingH := g.Spacing * float64(cols-1)
	cellW := (availW - totalSpacingH) / float64(cols)
	if cellW < 0 {
		cellW = 0
	}

	// Find uniform cell height from measured children
	maxCellH := 0.0
	var visible []*core.Node
	for _, child := range node.Children() {
		if child.GetVisibility() != core.Gone {
			visible = append(visible, child)
			sz := child.MeasuredSize()
			if sz.Height > maxCellH {
				maxCellH = sz.Height
			}
		}
	}

	for i, child := range visible {
		col := i % cols
		row := i / cols

		// 单元格左上角坐标
		cellX := padding.Left + float64(col)*(cellW+g.Spacing)
		cellY := padding.Top + float64(row)*(maxCellH+g.Spacing)

		// 子节点在单元格内偏移 margin
		margin := child.Margin()
		sz := child.MeasuredSize()
		x := cellX + margin.Left
		y := cellY + margin.Top

		child.SetBounds(core.Rect{X: x, Y: y, Width: sz.Width, Height: sz.Height})

		if l := child.GetLayout(); l != nil {
			l.Arrange(child, child.Bounds())
		}
	}
}
