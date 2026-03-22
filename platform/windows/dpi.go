package windows

// DpToPx converts density-independent pixels to physical pixels.
func DpToPx(dp, dpi float64) float64 {
	return dp * dpi / 96.0
}

// PxToDp converts physical pixels to density-independent pixels.
func PxToDp(px, dpi float64) float64 {
	return px * 96.0 / dpi
}
