package core

import "image/color"

// Style holds visual styling properties for a node.
// Will be expanded in Task 3.
type Style struct {
	BackgroundColor color.RGBA
	TextColor       color.RGBA
	FontSize        float64
	FontFamily      string
	FontWeight      int
}
