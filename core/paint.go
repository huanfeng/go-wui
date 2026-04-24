package core

import "image/color"

// PaintNodeRecursive 递归绘制节点及其子节点到画布上。
// 与 PaintNode 不同，此函数不处理 overlay 层 —— 适用于
// 需要自行管理子节点绘制的 widget（ScrollView、RecyclerView、SplitPane 等）。
// 它尊重节点的可见性、应用坐标平移，并通过 paintsChildren 标记
// 将子节点绘制权移交给节点自身的 Painter。
func PaintNodeRecursive(node *Node, canvas Canvas) {
	if node.GetVisibility() != Visible {
		return
	}
	canvas.Save()
	b := node.Bounds()
	canvas.Translate(b.X, b.Y)
	if painter := node.GetPainter(); painter != nil {
		painter.Paint(node, canvas)
	}
	// 如果节点自行处理子节点绘制（如 ScrollView），跳过递归子绘制
	if node.GetData("paintsChildren") == nil {
		for _, child := range node.Children() {
			PaintNodeRecursive(child, canvas)
		}
	}
	canvas.Restore()
}

// PaintNode 递归绘制节点树到画布上（完整重绘路径）。
// 它处理 overlay 子节点：带 "isOverlay" 标记的子节点会延迟到
// 所有普通子节点绘制完毕后再绘制，并撑满父节点尺寸，确保浮层
// 始终位于内容之上。
// 如果节点设置了 "paintsChildren" 数据标记，则跳过子节点递归，
// 由节点的 Painter 自行负责子节点绘制（如 ScrollView 的裁剪和滚动偏移）。
func PaintNode(node *Node, canvas Canvas) {
	if node.GetVisibility() != Visible {
		return
	}
	canvas.Save()
	b := node.Bounds()
	canvas.Translate(b.X, b.Y)
	if p := node.GetPainter(); p != nil {
		p.Paint(node, canvas)
	}
	// 如果 Painter 已经处理了子节点，跳过此处的递归绘制
	if node.GetData("paintsChildren") == nil {
		var overlays []*Node
		for _, child := range node.Children() {
			if child.GetData("isOverlay") != nil {
				overlays = append(overlays, child)
				continue
			}
			PaintNode(child, canvas)
		}
		// overlay 最后绘制，撑满父节点区域，覆盖所有内容
		for _, overlay := range overlays {
			overlay.SetBounds(Rect{X: 0, Y: 0, Width: b.Width, Height: b.Height})
			overlay.SetMeasuredSize(Size{Width: b.Width, Height: b.Height})
			PaintNode(overlay, canvas)
		}
	}
	canvas.Restore()
}

// PaintStyle controls whether shapes are filled, stroked, or both.
type PaintStyle int

const (
	PaintFill         PaintStyle = iota
	PaintStroke
	PaintFillAndStroke
)

// Paint holds drawing attributes used by Canvas operations.
type Paint struct {
	Color       color.RGBA
	DrawStyle   PaintStyle
	StrokeWidth float64
	FontSize    float64
	FontFamily  string
	FontWeight  int
	AntiAlias   bool
}
