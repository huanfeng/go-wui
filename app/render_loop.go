package app

import "gowui/core"

// PaintNode recursively paints a node tree onto a canvas.
// It respects visibility, applies translation for each node's bounds,
// delegates to the node's Painter, and recurses into children.
func PaintNode(node *core.Node, canvas core.Canvas) {
	if node.GetVisibility() != core.Visible {
		return
	}
	canvas.Save()
	b := node.Bounds()
	canvas.Translate(b.X, b.Y)
	if p := node.GetPainter(); p != nil {
		p.Paint(node, canvas)
	}
	for _, child := range node.Children() {
		PaintNode(child, canvas)
	}
	canvas.Restore()
}
