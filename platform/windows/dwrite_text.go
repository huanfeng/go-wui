//go:build windows

package windows

import (
	"gowui/core"
	"gowui/render/freetype"
)

// CreateTextRendererWithFallback returns DirectWrite if available, else FreeType.
func CreateTextRendererWithFallback() core.TextRenderer {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		return freetype.NewFreeTypeTextRenderer()
	}
	return tr
}
