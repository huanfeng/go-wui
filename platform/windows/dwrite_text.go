//go:build windows

package windows

import (
	"github.com/huanfeng/go-wui/core"
	"github.com/huanfeng/go-wui/render/freetype"
)

// CreateTextRendererWithFallback returns DirectWrite if available, else FreeType.
func CreateTextRendererWithFallback() core.TextRenderer {
	tr, err := NewDWriteTextRenderer()
	if err != nil {
		return freetype.NewFreeTypeTextRenderer()
	}
	return tr
}
