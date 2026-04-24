package widget

import (
	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/theme"
)

// ProgressBarStyle defines the visual style of a ProgressBar.
type ProgressBarStyle int

const (
	ProgressBarLinear   ProgressBarStyle = iota // Horizontal linear bar
	ProgressBarCircular                         // Circular (future)
)

// ProgressBar displays a linear progress indicator.
// Progress ranges from 0.0 to 1.0.
type ProgressBar struct {
	BaseView
	progress      float64 // 0.0 to 1.0
	indeterminate bool
}

const (
	progressBarHeight       = 4.0  // track height in dp
	progressBarCornerRadius = 2.0  // rounded ends
	progressBarIndeterminate = 0.3 // 30% width for indeterminate bar
)

// NewProgressBar creates a new linear ProgressBar with progress at 0.
func NewProgressBar() *ProgressBar {
	pb := &ProgressBar{}
	pb.node = initNode("ProgressBar", pb)
	pb.node.SetPainter(&progressBarPainter{pb: pb})
	pb.node.SetStyle(&core.Style{})
	pb.node.SetData("progress", 0.0)
	pb.node.SetData("indeterminate", false)
	return pb
}

// GetProgress returns the current progress value (0.0 to 1.0).
func (pb *ProgressBar) GetProgress() float64 {
	return pb.progress
}

// SetProgress sets the progress value, clamped to [0.0, 1.0].
func (pb *ProgressBar) SetProgress(progress float64) {
	if progress < 0.0 {
		progress = 0.0
	}
	if progress > 1.0 {
		progress = 1.0
	}
	pb.progress = progress
	pb.node.SetData("progress", progress)
	pb.node.MarkDirty()
}

// IsIndeterminate reports whether the progress bar is in indeterminate mode.
func (pb *ProgressBar) IsIndeterminate() bool {
	return pb.indeterminate
}

// SetIndeterminate sets whether the progress bar shows an indeterminate animation.
func (pb *ProgressBar) SetIndeterminate(indeterminate bool) {
	pb.indeterminate = indeterminate
	pb.node.SetData("indeterminate", indeterminate)
	pb.node.MarkDirty()
}

// progressBarPainter handles measurement and painting of the progress bar.
type progressBarPainter struct {
	pb *ProgressBar
}

func (p *progressBarPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Width: fill parent if exact, otherwise use spec size
	w := 200.0 // default width
	h := progressBarHeight

	switch ws.Mode {
	case core.MeasureModeExact:
		w = ws.Size
	case core.MeasureModeAtMost:
		if w > ws.Size {
			w = ws.Size
		}
	}

	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	} else if hs.Mode == core.MeasureModeAtMost && h > hs.Size {
		h = hs.Size
	}

	return core.Size{Width: w, Height: h}
}

func (p *progressBarPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()

	trackRect := core.Rect{
		X:      0,
		Y:      0,
		Width:  b.Width,
		Height: b.Height,
	}

	// Draw track background (使用主题分隔线色作为轨道背景)
	trackColor := theme.CurrentColors().Divider
	trackPaint := &core.Paint{
		Color:     trackColor,
		DrawStyle: core.PaintFill,
	}
	canvas.DrawRoundRect(trackRect, progressBarCornerRadius, trackPaint)

	// Draw progress fill (使用主题主色)
	primaryColor := theme.CurrentColors().Primary
	fillPaint := &core.Paint{
		Color:     primaryColor,
		DrawStyle: core.PaintFill,
	}

	var fillWidth float64
	if p.pb.indeterminate {
		// Indeterminate: fixed 30% width bar (static position; animation added later)
		fillWidth = b.Width * progressBarIndeterminate
	} else {
		fillWidth = b.Width * p.pb.progress
	}

	if fillWidth > 0 {
		fillRect := core.Rect{
			X:      0,
			Y:      0,
			Width:  fillWidth,
			Height: b.Height,
		}
		canvas.DrawRoundRect(fillRect, progressBarCornerRadius, fillPaint)
	}
}
