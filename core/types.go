package core

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Rect represents a rectangle with position and size.
type Rect struct {
	X, Y, Width, Height float64
}

// Contains reports whether the point (px, py) is inside the rectangle.
func (r Rect) Contains(px, py float64) bool {
	return px >= r.X && px <= r.X+r.Width &&
		py >= r.Y && py <= r.Y+r.Height
}

// Intersect returns the intersection of two rectangles.
// If they don't overlap, a zero-size Rect is returned.
func (r Rect) Intersect(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)

	w := max(0, x2-x1)
	h := max(0, y2-y1)

	return Rect{X: x1, Y: y1, Width: w, Height: h}
}

// ApplyInsets returns a new Rect shrunk by the given insets.
func (r Rect) ApplyInsets(insets Insets) Rect {
	return Rect{
		X:      r.X + insets.Left,
		Y:      r.Y + insets.Top,
		Width:  max(0, r.Width-insets.Left-insets.Right),
		Height: max(0, r.Height-insets.Top-insets.Bottom),
	}
}

// Size represents a width/height pair.
type Size struct {
	Width, Height float64
}

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

// Insets represents padding/margin on four sides.
type Insets struct {
	Left, Top, Right, Bottom float64
}

// DimensionUnit enumerates how a Dimension value is interpreted.
type DimensionUnit int

const (
	DimensionPx           DimensionUnit = iota // absolute pixels
	DimensionDp                                // density-independent pixels
	DimensionMatchParent                       // fill parent
	DimensionWrapContent                       // shrink to content
	DimensionWeight                            // weighted flex
)

// Dimension holds a numeric value and its unit.
type Dimension struct {
	Value float64
	Unit  DimensionUnit
}

// String returns a human-readable representation for debugging.
func (d Dimension) String() string {
	switch d.Unit {
	case DimensionPx:
		return fmt.Sprintf("%.0fpx", d.Value)
	case DimensionDp:
		return fmt.Sprintf("%.0fdp", d.Value)
	case DimensionMatchParent:
		return "match_parent"
	case DimensionWrapContent:
		return "wrap_content"
	case DimensionWeight:
		return fmt.Sprintf("%.1fw", d.Value)
	default:
		return fmt.Sprintf("%.0f?", d.Value)
	}
}

// ParseDimension parses a dimension string such as "200dp", "100px",
// "match_parent", or "wrap_content".
func ParseDimension(s string) Dimension {
	switch s {
	case "match_parent":
		return Dimension{Unit: DimensionMatchParent}
	case "wrap_content":
		return Dimension{Unit: DimensionWrapContent}
	}

	if strings.HasSuffix(s, "dp") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "dp"), 64)
		return Dimension{Value: v, Unit: DimensionDp}
	}
	if strings.HasSuffix(s, "px") {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		return Dimension{Value: v, Unit: DimensionPx}
	}

	// Fallback: try parsing as plain number (pixels)
	v, _ := strconv.ParseFloat(s, 64)
	return Dimension{Value: v, Unit: DimensionPx}
}

// ParseColor parses a hex color string in "#RRGGBB" or "#AARRGGBB" format.
func ParseColor(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 6: // RRGGBB
		r, _ := strconv.ParseUint(s[0:2], 16, 8)
		g, _ := strconv.ParseUint(s[2:4], 16, 8)
		b, _ := strconv.ParseUint(s[4:6], 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
	case 8: // AARRGGBB
		a, _ := strconv.ParseUint(s[0:2], 16, 8)
		r, _ := strconv.ParseUint(s[2:4], 16, 8)
		g, _ := strconv.ParseUint(s[4:6], 16, 8)
		b, _ := strconv.ParseUint(s[6:8], 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	default:
		return color.RGBA{}
	}
}
