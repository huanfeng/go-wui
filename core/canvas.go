package core

import "image"

// Canvas provides platform-independent 2D drawing.
type Canvas interface {
	DrawRect(rect Rect, paint *Paint)
	DrawRoundRect(rect Rect, radius float64, paint *Paint)
	DrawCircle(cx, cy, radius float64, paint *Paint)
	DrawLine(x1, y1, x2, y2 float64, paint *Paint)
	DrawImage(img *ImageResource, dst Rect)
	DrawText(text string, x, y float64, paint *Paint)
	MeasureText(text string, paint *Paint) Size

	Save()
	Restore()
	Translate(dx, dy float64)
	ClipRect(rect Rect)

	Target() *image.RGBA
}
