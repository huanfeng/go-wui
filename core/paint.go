package core

import "image/color"

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
