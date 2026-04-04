//go:build windows

package windows

import (
	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/render/freetype"
)

// CreateTextRendererWithFallback returns DirectWrite if available, else FreeType.
func CreateTextRendererWithFallback() core.TextRenderer {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		return freetype.NewFreeTypeTextRenderer()
	}
	return tr
}
