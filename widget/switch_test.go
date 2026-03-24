package widget

import (
	"github.com/huanfeng/go-wui/core"
	"testing"
)

func TestSwitchToggle(t *testing.T) {
	sw := NewSwitch()
	if sw.IsOn() {
		t.Error("should start off")
	}
	sw.SetOn(true)
	if !sw.IsOn() {
		t.Error("should be on")
	}
	sw.SetOn(false)
	if sw.IsOn() {
		t.Error("should be off after toggling back")
	}
}

func TestSwitchOnChanged(t *testing.T) {
	sw := NewSwitch()
	var received bool
	sw.SetOnChanged(func(on bool) { received = on })

	// Simulate click via handler
	handler := sw.Node().GetHandler()
	sw.Node().SetBounds(core.Rect{Width: 44, Height: 24})

	down := core.NewMotionEvent(core.ActionDown, 22, 12)
	handler.OnEvent(sw.Node(), down)

	up := core.NewMotionEvent(core.ActionUp, 22, 12)
	handler.OnEvent(sw.Node(), up)

	if !received {
		t.Error("onChanged should have been called with true")
	}
	if !sw.IsOn() {
		t.Error("switch should be on after click")
	}
}

func TestSwitchDoubleClick(t *testing.T) {
	sw := NewSwitch()
	handler := sw.Node().GetHandler()
	sw.Node().SetBounds(core.Rect{Width: 44, Height: 24})

	// First click: off -> on
	handler.OnEvent(sw.Node(), core.NewMotionEvent(core.ActionDown, 22, 12))
	handler.OnEvent(sw.Node(), core.NewMotionEvent(core.ActionUp, 22, 12))
	if !sw.IsOn() {
		t.Error("should be on after first click")
	}

	// Second click: on -> off
	handler.OnEvent(sw.Node(), core.NewMotionEvent(core.ActionDown, 22, 12))
	handler.OnEvent(sw.Node(), core.NewMotionEvent(core.ActionUp, 22, 12))
	if sw.IsOn() {
		t.Error("should be off after second click")
	}
}
