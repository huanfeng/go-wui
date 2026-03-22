package core

import (
	"math"
	"time"
)

// Interpolator maps an input fraction (0.0-1.0) to an output value.
type Interpolator interface {
	GetInterpolation(fraction float64) float64
}

// LinearInterpolator — constant rate.
type LinearInterpolator struct{}

func (i *LinearInterpolator) GetInterpolation(f float64) float64 { return f }

// AccelerateDecelerateInterpolator — slow start, accelerate, then decelerate.
type AccelerateDecelerateInterpolator struct{}

func (i *AccelerateDecelerateInterpolator) GetInterpolation(f float64) float64 {
	return (math.Cos((f+1)*math.Pi)/2.0) + 0.5
}

// DecelerateInterpolator — fast start, then decelerate.
type DecelerateInterpolator struct{}

func (i *DecelerateInterpolator) GetInterpolation(f float64) float64 {
	return 1.0 - (1.0-f)*(1.0-f)
}

// ValueAnimator drives a float64 value from From to To over Duration.
type ValueAnimator struct {
	From     float64
	To       float64
	Duration time.Duration
	Interp   Interpolator
	OnUpdate func(value float64)
	OnEnd    func()

	startTime time.Duration // accumulated time
	running   bool
	finished  bool
}

// Start begins (or restarts) the animation.
func (a *ValueAnimator) Start() {
	a.startTime = 0
	a.running = true
	a.finished = false
}

// Tick advances the animation by elapsed time since last tick.
// Returns true if still running.
func (a *ValueAnimator) Tick(elapsed time.Duration) bool {
	if !a.running || a.finished {
		return false
	}
	a.startTime += elapsed
	fraction := float64(a.startTime) / float64(a.Duration)
	if fraction >= 1.0 {
		fraction = 1.0
		a.finished = true
		a.running = false
	}
	interp := a.Interp
	if interp == nil {
		interp = &LinearInterpolator{}
	}
	interpolated := interp.GetInterpolation(fraction)
	value := a.From + (a.To-a.From)*interpolated
	if a.OnUpdate != nil {
		a.OnUpdate(value)
	}
	if a.finished && a.OnEnd != nil {
		a.OnEnd()
	}
	return !a.finished
}

// IsFinished reports whether the animation has completed.
func (a *ValueAnimator) IsFinished() bool { return a.finished }

// IsRunning reports whether the animation is currently active.
func (a *ValueAnimator) IsRunning() bool { return a.running }

// Cancel stops the animation immediately.
func (a *ValueAnimator) Cancel() {
	a.running = false
	a.finished = true
}
