package core

// Painter handles measurement and rendering of node content.
// Canvas is defined in the same package (core/canvas.go will be added later in Task 5).
// For now, use a forward-compatible interface parameter.
type Painter interface {
	Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
	Paint(node *Node, canvas interface{})
}
