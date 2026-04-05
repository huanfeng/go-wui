package platform

import "github.com/huanfeng/wind-ui/core"

// WindowOptions configures a new window before it is created.
type WindowOptions struct {
	Title       string
	Width       int
	Height      int
	MinWidth    int
	MinHeight   int
	X, Y        int
	Resizable   bool
	Frameless   bool
	TopMost     bool
	Transparent bool
	NoActivate  bool // window does not steal focus when shown
	Icon        *core.ImageResource
}

// Window is the platform-independent abstraction for a native OS window.
type Window interface {
	SetContentView(root *core.Node)
	SetTitle(title string)
	SetIcon(icon *core.ImageResource)
	Show()
	Hide()
	Close()
	Minimize()
	Maximize()
	Restore()
	SetSize(width, height int)
	SetPosition(x, y int)
	Center()
	IsVisible() bool
	IsFocused() bool
	GetSize() core.Size
	GetPosition() core.Point
	GetDPI() float64
	SetOnClose(fn func() bool)
	SetOnResize(fn func(w, h int))
	SetOnDPIChanged(fn func(dpi float64))
	SetOnFocusChanged(fn func(focused bool))
	NativeHandle() uintptr
	Invalidate()
	InvalidateRect(rect core.Rect)
	StartAnimator(anim *core.ValueAnimator)
	RequestFrame()
}

// NativeEditText wraps a platform-native text input control.
type NativeEditText interface {
	AttachToNode(node *core.Node)
	Detach()
	GetText() string
	SetText(text string)
	SetPlaceholder(text string)
	SetFont(family string, size float64, weight int)
	SetTextColor(clr interface{})
	SetBackgroundColor(clr interface{})
	SetMultiLine(multiLine bool)
	SetMaxLength(max int)
	SetInputType(inputType InputType)
	SetOnTextChanged(fn func(text string))
	SetOnSubmit(fn func(text string))
	Focus()
	ClearFocus()
}
