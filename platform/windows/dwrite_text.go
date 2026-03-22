package windows

import (
	"gowui/core"
	"gowui/render/freetype"
)

// NewDWriteTextRenderer attempts to load the DirectWrite DLL.
// If not available, returns nil and the caller should fall back to FreeType.
func NewDWriteTextRenderer() core.TextRenderer {
	// Phase 1: DirectWrite DLL not yet built, always return nil.
	// Phase 2: Load gowui_dwrite.dll and wrap it.
	return nil
}

// CreateTextRendererWithFallback returns DirectWrite if available, else FreeType.
func CreateTextRendererWithFallback() core.TextRenderer {
	if tr := NewDWriteTextRenderer(); tr != nil {
		return tr
	}
	return freetype.NewFreeTypeTextRenderer()
}
