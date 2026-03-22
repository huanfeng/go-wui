package core

// Painter handles measurement and rendering of node content.
type Painter interface {
	Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
	Paint(node *Node, canvas Canvas)
}
