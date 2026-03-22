package app

import "gowui/core"

// PaintNode recursively paints a node tree onto a canvas.
// It respects visibility, applies translation for each node's bounds,
// delegates to the node's Painter, and recurses into children.
//
// If a node has the "paintsChildren" data flag set, child painting is
// skipped here because the node's Painter handles it internally (e.g.
// ScrollView applies clipping and scroll-offset translation).
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
	// Skip child painting if the painter already handled it
	if node.GetData("paintsChildren") == nil {
		for _, child := range node.Children() {
			PaintNode(child, canvas)
		}
	}
	canvas.Restore()
}
