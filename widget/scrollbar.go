package widget

import (
	"image/color"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
)

// Scrollbar is a reusable scrollbar component that handles painting and
// interaction state. It can be embedded in ScrollView, RecyclerView,
// TreeView, or any scrollable container.
type Scrollbar struct {
	Orientation layout.Orientation
	Hovered     bool
	Dragging    bool
	dragAnchor  float64 // offset from thumb top to the grab point
}

// ScrollbarMetrics holds computed thumb position and size.
type ScrollbarMetrics struct {
	ThumbPos    float64 // thumb offset along the scroll axis
	ThumbLen    float64 // thumb length
	ScrollRange float64 // max scroll offset
	TrackLen    float64 // available track length (viewport - thumb)
	HasScroll   bool    // true if content exceeds viewport
}

const (
	sbWidth     = 8.0  // scrollbar width (consistent in all states)
	sbMinThumb  = 24.0 // minimum thumb length
	sbMargin    = 2.0  // margin from edge
	sbHitExtra  = 6.0  // extra hit area for easier grabbing
)

// Scrollbar colors — premultiplied-alpha safe (R,G,B ≤ A).
var (
	sbTrackColor      = color.RGBA{R: 20, G: 20, B: 20, A: 25}    // barely visible
	sbThumbDefault    = color.RGBA{R: 100, G: 100, B: 100, A: 140} // normal
	sbThumbHovered    = color.RGBA{R: 90, G: 90, B: 90, A: 180}    // darker on hover
	sbThumbDragging   = color.RGBA{R: 80, G: 80, B: 80, A: 210}    // darkest on drag
)

// ComputeMetrics calculates the scrollbar thumb position and size.
func (sb *Scrollbar) ComputeMetrics(viewportSize, contentSize, scrollOffset float64) ScrollbarMetrics {
	if contentSize <= viewportSize || contentSize == 0 {
		return ScrollbarMetrics{}
	}

	ratio := viewportSize / contentSize
	thumbLen := viewportSize * ratio
	if thumbLen < sbMinThumb {
		thumbLen = sbMinThumb
	}

	scrollRange := contentSize - viewportSize
	if scrollRange <= 0 {
		return ScrollbarMetrics{}
	}

	trackLen := viewportSize - thumbLen
	fraction := scrollOffset / scrollRange
	thumbPos := trackLen * fraction

	return ScrollbarMetrics{
		ThumbPos:    thumbPos,
		ThumbLen:    thumbLen,
		ScrollRange: scrollRange,
		TrackLen:    trackLen,
		HasScroll:   true,
	}
}

// Paint draws the scrollbar (track + thumb) on the canvas.
// bounds is the local rect of the parent scrollable container.
func (sb *Scrollbar) Paint(canvas core.Canvas, bounds core.Rect, metrics ScrollbarMetrics) {
	if !metrics.HasScroll {
		return
	}

	thumbColor := sbThumbDefault
	if sb.Dragging {
		thumbColor = sbThumbDragging
	} else if sb.Hovered {
		thumbColor = sbThumbHovered
	}

	trackPaint := &core.Paint{Color: sbTrackColor, DrawStyle: core.PaintFill}
	thumbPaint := &core.Paint{Color: thumbColor, DrawStyle: core.PaintFill}

	if sb.Orientation == layout.Vertical {
		trackX := bounds.Width - sbWidth - sbMargin
		canvas.DrawRect(core.Rect{X: trackX, Y: 0, Width: sbWidth, Height: bounds.Height}, trackPaint)
		canvas.DrawRoundRect(core.Rect{
			X: trackX, Y: metrics.ThumbPos, Width: sbWidth, Height: metrics.ThumbLen,
		}, sbWidth/2, thumbPaint)
	} else {
		trackY := bounds.Height - sbWidth - sbMargin
		canvas.DrawRect(core.Rect{X: 0, Y: trackY, Width: bounds.Width, Height: sbWidth}, trackPaint)
		canvas.DrawRoundRect(core.Rect{
			X: metrics.ThumbPos, Y: trackY, Width: metrics.ThumbLen, Height: sbWidth,
		}, sbWidth/2, thumbPaint)
	}
}

// IsInHitArea checks if the given local coordinates are in the scrollbar area.
func (sb *Scrollbar) IsInHitArea(localX, localY, boundsW, boundsH float64) bool {
	hitW := sbWidth + sbMargin + sbHitExtra
	if sb.Orientation == layout.Vertical {
		return localX >= boundsW-hitW && localX <= boundsW && localY >= 0 && localY <= boundsH
	}
	return localY >= boundsH-hitW && localY <= boundsH && localX >= 0 && localX <= boundsW
}

// isOnThumb checks if a position along the scroll axis is on the thumb.
func (sb *Scrollbar) isOnThumb(pos float64, metrics ScrollbarMetrics) bool {
	return pos >= metrics.ThumbPos && pos <= metrics.ThumbPos+metrics.ThumbLen
}

// ScrollToPosition returns the scroll offset for a position along the scroll axis.
// anchor is the offset from the thumb's top edge to the grab point.
func (sb *Scrollbar) ScrollToPosition(localPos float64, metrics ScrollbarMetrics) float64 {
	if metrics.TrackLen <= 0 || metrics.ScrollRange <= 0 {
		return 0
	}
	fraction := (localPos - sb.dragAnchor) / metrics.TrackLen
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return fraction * metrics.ScrollRange
}

// HandleEvent processes a motion event for the scrollbar.
// node is the scrollable container's node (for AbsolutePosition).
// Returns (consumed, newScrollOffset). Only valid if consumed is true.
func (sb *Scrollbar) HandleEvent(node *core.Node, event core.Event, metrics ScrollbarMetrics, currentOffset float64) (consumed bool, newOffset float64) {
	me, ok := event.(*core.MotionEvent)
	if !ok {
		return false, currentOffset
	}

	pos := node.AbsolutePosition()
	b := node.Bounds()
	localX := me.X - pos.X
	localY := me.Y - pos.Y

	switch me.Action {
	case core.ActionDown:
		if sb.IsInHitArea(localX, localY, b.Width, b.Height) {
			sb.Dragging = true
			sb.Hovered = true

			// Check if clicking on the thumb itself vs the track
			var scrollPos float64
			if sb.Orientation == layout.Vertical {
				scrollPos = localY
			} else {
				scrollPos = localX
			}

			if sb.isOnThumb(scrollPos, metrics) {
				// Clicking on thumb: start drag without jumping.
				// Store the offset between click position and thumb top
				// so dragging feels anchored to where the user grabbed.
				sb.dragAnchor = scrollPos - metrics.ThumbPos
				newOffset = currentOffset // no change
			} else {
				// Clicking on track: jump to center thumb at click position
				sb.dragAnchor = metrics.ThumbLen / 2
				newOffset = sb.ScrollToPosition(scrollPos, metrics)
			}
			return true, newOffset
		}

	case core.ActionMove:
		if sb.Dragging {
			var scrollPos float64
			if sb.Orientation == layout.Vertical {
				scrollPos = localY
			} else {
				scrollPos = localX
			}
			// Use dragAnchor so thumb follows the grab point smoothly
			newOffset = sb.ScrollToPosition(scrollPos, metrics)
			return true, newOffset
		}

	case core.ActionUp:
		if sb.Dragging {
			sb.Dragging = false
			// Keep hovered if mouse is still over the scrollbar
			sb.Hovered = sb.IsInHitArea(localX, localY, b.Width, b.Height)
			return true, currentOffset
		}

	case core.ActionHoverEnter, core.ActionHoverMove:
		wasHovered := sb.Hovered
		sb.Hovered = sb.IsInHitArea(localX, localY, b.Width, b.Height)
		if sb.Hovered != wasHovered {
			node.MarkDirty()
		}
		return sb.Hovered, currentOffset

	case core.ActionHoverExit:
		if sb.Hovered {
			sb.Hovered = false
			node.MarkDirty()
		}
		return false, currentOffset

	case core.ActionCancel:
		sb.Dragging = false
		sb.Hovered = false
		return true, currentOffset
	}

	return false, currentOffset
}

// ClearHover resets hover state (call when receiving scroll wheel events).
func (sb *Scrollbar) ClearHover() {
	sb.Hovered = false
}
