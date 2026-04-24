package core

// TextRenderer handles text measurement and rendering.
type TextRenderer interface {
	SetFont(fontFamily string, weight int, size float64)
	MeasureText(text string) Size
	DrawText(canvas Canvas, text string, x, y float64, paint *Paint)
	CreateTextLayout(text string, paint *Paint, maxWidth float64) *TextLayoutResult
	Close()
}

// TextMeasurer provides text measurement during layout.
// Stored on the root node (key "textMeasurer") so that Painter.Measure()
// implementations can obtain accurate text dimensions instead of estimates.
type TextMeasurer interface {
	MeasureText(text string, paint *Paint) Size
}

// textMeasurerAdapter wraps a TextRenderer to satisfy TextMeasurer.
type textMeasurerAdapter struct {
	tr TextRenderer
}

func (a *textMeasurerAdapter) MeasureText(text string, paint *Paint) Size {
	if a.tr == nil || text == "" {
		return Size{}
	}
	if paint != nil {
		a.tr.SetFont(paint.FontFamily, paint.FontWeight, paint.FontSize)
	}
	return a.tr.MeasureText(text)
}

// NewTextMeasurer wraps a TextRenderer into a TextMeasurer suitable for
// storing on the root node via SetData("textMeasurer", ...).
func NewTextMeasurer(tr TextRenderer) TextMeasurer {
	return &textMeasurerAdapter{tr: tr}
}

// GetTextMeasurer walks up the node tree to find a TextMeasurer stored
// on an ancestor (typically the root node). Returns nil if none is found.
func GetTextMeasurer(node *Node) TextMeasurer {
	for n := node; n != nil; n = n.parent {
		if tm, ok := n.GetData("textMeasurer").(TextMeasurer); ok {
			return tm
		}
	}
	return nil
}

// NodeMeasureText measures text using the TextMeasurer found in the node tree.
// If no TextMeasurer is available, it falls back to a rough character-width estimate.
func NodeMeasureText(node *Node, text string, paint *Paint) Size {
	if text == "" {
		return Size{}
	}
	if tm := GetTextMeasurer(node); tm != nil {
		return tm.MeasureText(text, paint)
	}
	// Fallback: rough estimate for environments without a TextMeasurer
	fontSize := 14.0
	if paint != nil && paint.FontSize > 0 {
		fontSize = paint.FontSize
	}
	charWidth := fontSize * 0.6
	return Size{
		Width:  float64(len([]rune(text))) * charWidth,
		Height: fontSize * 1.4,
	}
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
