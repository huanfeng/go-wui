package core

import "image"

// ImageResource holds a decoded image and its metadata.
type ImageResource struct {
	Image  *image.RGBA
	Width  int
	Height int
	Name   string
}
