package core

// MeasureMode defines how the parent constrains a child's size.
type MeasureMode int

const (
	MeasureModeExact   MeasureMode = iota // Exact size (dp value or match_parent resolved by parent)
	MeasureModeAtMost                     // Up to this size (wrap_content)
	MeasureModeUnbound                    // No constraint (ScrollView children)
)

// MeasureSpec combines a MeasureMode with a size constraint.
type MeasureSpec struct {
	Mode MeasureMode
	Size float64
}

// Layout defines a strategy for measuring and arranging child nodes.
type Layout interface {
	Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
	Arrange(node *Node, bounds Rect)
}

// DPIScalable is optionally implemented by Layout types that hold
// dp-valued fields (spacing, etc.) which need DPI scaling.
type DPIScalable interface {
	ScaleDPI(scale float64)
}
