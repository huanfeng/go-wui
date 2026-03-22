package core

// TextRenderer handles text measurement and rendering.
type TextRenderer interface {
	SetFont(fontFamily string, weight int, size float64)
	MeasureText(text string) Size
	DrawText(canvas Canvas, text string, x, y float64, paint *Paint)
	CreateTextLayout(text string, paint *Paint, maxWidth float64) *TextLayoutResult
	Close()
}

// TextLayoutResult holds the result of multi-line text layout.
type TextLayoutResult struct {
	Lines     []TextLine
	TotalSize Size
}

// TextLine describes a single line within a text layout.
type TextLine struct {
	Text     string
	Offset   Point
	Width    float64
	Baseline float64
}
