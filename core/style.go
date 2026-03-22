package core

import "image/color"

// Style holds visual styling properties for a node.
type Style struct {
	BackgroundColor color.RGBA
	BorderColor     color.RGBA
	BorderWidth     float64
	CornerRadius    float64
	FontSize        float64
	FontFamily      string
	FontWeight      int
	TextColor       color.RGBA
	Opacity         float64

	Width  Dimension
	Height Dimension
	Weight float64 // layout_weight for LinearLayout

	Gravity     Gravity
	TextGravity Gravity
}

// Gravity controls the alignment of content within its container.
type Gravity int

const (
	GravityStart            Gravity = 0
	GravityCenter           Gravity = 1
	GravityEnd              Gravity = 2
	GravityCenterVertical   Gravity = 4
	GravityCenterHorizontal Gravity = 8
)

// MergeStyles creates a new style with child values overriding parent.
// Zero values in child mean "not set" — inherited from parent.
func MergeStyles(parent, child *Style) *Style {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent
	if child.BackgroundColor != (color.RGBA{}) {
		merged.BackgroundColor = child.BackgroundColor
	}
	if child.BorderColor != (color.RGBA{}) {
		merged.BorderColor = child.BorderColor
	}
	if child.BorderWidth != 0 {
		merged.BorderWidth = child.BorderWidth
	}
	if child.CornerRadius != 0 {
		merged.CornerRadius = child.CornerRadius
	}
	if child.FontSize != 0 {
		merged.FontSize = child.FontSize
	}
	if child.FontFamily != "" {
		merged.FontFamily = child.FontFamily
	}
	if child.FontWeight != 0 {
		merged.FontWeight = child.FontWeight
	}
	if child.TextColor != (color.RGBA{}) {
		merged.TextColor = child.TextColor
	}
	if child.Opacity != 0 {
		merged.Opacity = child.Opacity
	}
	return &merged
}
