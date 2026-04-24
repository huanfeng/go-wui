package theme

import "image/color"

// DefaultColors 定义主题的默认颜色方案
type DefaultColors struct {
	Primary       color.RGBA // 主色（按钮背景等）
	PrimaryDark   color.RGBA // 深主色
	Accent        color.RGBA // 强调色
	Background    color.RGBA // 窗口背景
	Surface       color.RGBA // 卡片/容器背景
	TextPrimary   color.RGBA // 主要文本颜色
	TextSecondary color.RGBA // 次要文本颜色
	TextOnPrimary color.RGBA // 主色背景上的文本颜色
	Divider       color.RGBA // 分隔线颜色
	Border        color.RGBA // 边框颜色
	Disabled      color.RGBA // 禁用状态颜色
}

// LightColors 是浅色主题的默认颜色方案
var LightColors = DefaultColors{
	Primary:       color.RGBA{R: 25, G: 118, B: 210, A: 255},  // #1976D2
	PrimaryDark:   color.RGBA{R: 21, G: 101, B: 192, A: 255},  // #1565C0
	Accent:        color.RGBA{R: 255, G: 87, B: 34, A: 255},   // #FF5722
	Background:    color.RGBA{R: 250, G: 250, B: 250, A: 255}, // #FAFAFA
	Surface:       color.RGBA{R: 255, G: 255, B: 255, A: 255}, // #FFFFFF
	TextPrimary:   color.RGBA{R: 33, G: 33, B: 33, A: 255},    // #212121
	TextSecondary: color.RGBA{R: 117, G: 117, B: 117, A: 255}, // #757575
	TextOnPrimary: color.RGBA{R: 255, G: 255, B: 255, A: 255}, // #FFFFFF
	Divider:       color.RGBA{R: 224, G: 224, B: 224, A: 255}, // #E0E0E0
	Border:        color.RGBA{R: 200, G: 200, B: 200, A: 255}, // #C8C8C8
	Disabled:      color.RGBA{R: 189, G: 189, B: 189, A: 255}, // #BDBDBD
}

// DarkColors 是深色主题的默认颜色方案
var DarkColors = DefaultColors{
	Primary:       color.RGBA{R: 100, G: 181, B: 246, A: 255}, // #64B5F6
	PrimaryDark:   color.RGBA{R: 66, G: 165, B: 245, A: 255},  // #42A5F5
	Accent:        color.RGBA{R: 255, G: 138, B: 101, A: 255}, // #FF8A65
	Background:    color.RGBA{R: 18, G: 18, B: 18, A: 255},    // #121212
	Surface:       color.RGBA{R: 30, G: 30, B: 30, A: 255},    // #1E1E1E
	TextPrimary:   color.RGBA{R: 255, G: 255, B: 255, A: 255}, // #FFFFFF
	TextSecondary: color.RGBA{R: 176, G: 176, B: 176, A: 255}, // #B0B0B0
	TextOnPrimary: color.RGBA{R: 0, G: 0, B: 0, A: 255},       // #000000
	Divider:       color.RGBA{R: 66, G: 66, B: 66, A: 255},    // #424242
	Border:        color.RGBA{R: 80, G: 80, B: 80, A: 255},    // #505050
	Disabled:      color.RGBA{R: 100, G: 100, B: 100, A: 255}, // #646464
}

// currentColors 存储当前激活的颜色方案，默认为浅色
var currentColors = LightColors

// CurrentColors 返回当前主题的颜色方案
func CurrentColors() DefaultColors {
	return currentColors
}

// SetTheme 切换主题颜色（dark=true 切换为深色主题，false 切换为浅色主题）
func SetTheme(dark bool) {
	if dark {
		currentColors = DarkColors
	} else {
		currentColors = LightColors
	}
}

// IsDark 返回当前是否为深色主题
func IsDark() bool {
	return currentColors.Background.R < 128
}
