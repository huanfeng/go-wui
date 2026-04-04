package widget

import "github.com/huanfeng/wind-ui/core"

// ScaleType controls how an image is scaled within its bounds.
type ScaleType int

const (
	ScaleFitCenter  ScaleType = iota // Scale to fit, centered
	ScaleCenterCrop                  // Scale to fill, crop excess
	ScaleFitXY                       // Stretch to fill exactly
)

// ImageView displays an image resource.
type ImageView struct {
	BaseView
	scaleType ScaleType
}

// NewImageView creates a new ImageView with default ScaleFitCenter.
func NewImageView() *ImageView {
	iv := &ImageView{scaleType: ScaleFitCenter}
	iv.node = initNode("ImageView", iv)
	iv.node.SetPainter(&imageViewPainter{iv: iv})
	iv.node.SetStyle(&core.Style{})
	return iv
}

// SetImage sets the image resource to display.
func (iv *ImageView) SetImage(img *core.ImageResource) {
	iv.node.SetData("image", img)
	iv.node.MarkDirty()
}

// SetScaleType sets how the image is scaled within bounds.
func (iv *ImageView) SetScaleType(st ScaleType) {
	iv.scaleType = st
}

// imageViewPainter measures and draws the image.
type imageViewPainter struct {
	iv *ImageView
}

func (p *imageViewPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	img, _ := node.GetData("image").(*core.ImageResource)
	w, h := 0.0, 0.0
	if img != nil {
		w = float64(img.Width)
		h = float64(img.Height)
	}
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if ws.Mode == core.MeasureModeAtMost && w > ws.Size {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	if hs.Mode == core.MeasureModeAtMost && h > hs.Size {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *imageViewPainter) Paint(node *core.Node, canvas core.Canvas) {
	img, _ := node.GetData("image").(*core.ImageResource)
	if img == nil {
		return
	}
	b := node.Bounds()
	canvas.DrawImage(img, core.Rect{Width: b.Width, Height: b.Height})
}
