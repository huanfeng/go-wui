package core

import (
	"math"
	"testing"
	"time"
)

func TestLinearInterpolator(t *testing.T) {
	li := &LinearInterpolator{}
	if li.GetInterpolation(0) != 0 {
		t.Error("0")
	}
	if li.GetInterpolation(0.5) != 0.5 {
		t.Error("0.5")
	}
	if li.GetInterpolation(1.0) != 1.0 {
		t.Error("1.0")
	}
}

func TestDecelerateInterpolator(t *testing.T) {
	di := &DecelerateInterpolator{}
	// At 0.5 input, decelerate should be > 0.5 (faster early)
	val := di.GetInterpolation(0.5)
	if val <= 0.5 {
		t.Errorf("decelerate at 0.5 should be > 0.5, got %v", val)
	}
	if di.GetInterpolation(0) != 0 {
		t.Error("0")
	}
	if di.GetInterpolation(1.0) != 1.0 {
		t.Error("1.0")
	}
}

func TestAccelerateDecelerateInterpolator(t *testing.T) {
	adi := &AccelerateDecelerateInterpolator{}
	if math.Abs(adi.GetInterpolation(0)) > 0.001 {
		t.Error("should start near 0")
	}
	if math.Abs(adi.GetInterpolation(1.0)-1.0) > 0.001 {
		t.Error("should end near 1")
	}
	// At midpoint should be ~0.5
	mid := adi.GetInterpolation(0.5)
	if math.Abs(mid-0.5) > 0.01 {
		t.Errorf("midpoint should be ~0.5, got %v", mid)
	}
}

func TestValueAnimator(t *testing.T) {
	var values []float64
	anim := &ValueAnimator{
		From:     0,
		To:       100,
		Duration: 100 * time.Millisecond,
		Interp:   &LinearInterpolator{},
		OnUpdate: func(v float64) { values = append(values, v) },
	}
	anim.Start()
	anim.Tick(50 * time.Millisecond)  // 50% → 50
	anim.Tick(50 * time.Millisecond)  // 100% → 100

	if len(values) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(values))
	}
	if values[0] != 50 {
		t.Errorf("first value: got %v, want 50", values[0])
	}
	if values[1] != 100 {
		t.Errorf("last value: got %v, want 100", values[1])
	}
	if !anim.IsFinished() {
		t.Error("should be finished")
	}
}

func TestValueAnimatorOnEnd(t *testing.T) {
	ended := false
	anim := &ValueAnimator{
		From:     0,
		To:       1,
		Duration: 10 * time.Millisecond,
		OnEnd:    func() { ended = true },
	}
	anim.Start()
	anim.Tick(20 * time.Millisecond) // overshoot
	if !ended {
		t.Error("OnEnd should have been called")
	}
}

func TestValueAnimatorCancel(t *testing.T) {
	anim := &ValueAnimator{
		From:     0,
		To:       100,
		Duration: 100 * time.Millisecond,
		OnUpdate: func(v float64) {},
	}
	anim.Start()
	anim.Tick(30 * time.Millisecond)
	anim.Cancel()
	if !anim.IsFinished() {
		t.Error("should be finished after cancel")
	}
	if anim.IsRunning() {
		t.Error("should not be running after cancel")
	}
	// Tick after cancel should be no-op
	result := anim.Tick(10 * time.Millisecond)
	if result {
		t.Error("tick after cancel should return false")
	}
}
