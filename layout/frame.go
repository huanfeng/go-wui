package layout

import "gowui/core"

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

		style := child.GetStyle()
		childWidthSpec := childMeasureSpec(widthSpec, paddingH, dimOrDefault(style, true))
		childHeightSpec := childMeasureSpec(heightSpec, paddingV, dimOrDefault(style, false))

		sz := MeasureChild(child, childWidthSpec, childHeightSpec)
		if sz.Width > maxWidth {
			maxWidth = sz.Width
		}
		if sz.Height > maxHeight {
			maxHeight = sz.Height
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
	contentX := bounds.X + padding.Left
	contentY := bounds.Y + padding.Top
	contentW := bounds.Width - padding.Left - padding.Right
	contentH := bounds.Height - padding.Top - padding.Bottom

	for _, child := range node.Children() {
		if child.GetVisibility() == core.Gone {
			continue
		}

		sz := child.MeasuredSize()
		style := child.GetStyle()

		gravity := core.GravityStart
		if style != nil {
			gravity = style.Gravity
		}

		var x, y float64

		switch gravity {
		case core.GravityCenter:
			x = contentX + (contentW-sz.Width)/2
			y = contentY + (contentH-sz.Height)/2
		case core.GravityEnd:
			x = contentX + contentW - sz.Width
			y = contentY + contentH - sz.Height
		default: // GravityStart
			x = contentX
			y = contentY
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
