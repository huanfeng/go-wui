package widget

import (
	"gowui/core"
	"testing"
)

func TestCheckBoxToggle(t *testing.T) {
	cb := NewCheckBox("Accept terms")
	if cb.IsChecked() {
		t.Error("should start unchecked")
	}
	cb.SetChecked(true)
	if !cb.IsChecked() {
		t.Error("should be checked")
	}
	cb.SetChecked(false)
	if cb.IsChecked() {
		t.Error("should be unchecked after toggling back")
	}
}

func TestCheckBoxOnChanged(t *testing.T) {
	cb := NewCheckBox("Test")
	var received bool
	cb.SetOnCheckedChanged(func(checked bool) { received = checked })

	// Simulate click via handler
	handler := cb.Node().GetHandler()
	cb.Node().SetBounds(core.Rect{Width: 200, Height: 30})

	down := core.NewMotionEvent(core.ActionDown, 10, 15)
	handler.OnEvent(cb.Node(), down)

	up := core.NewMotionEvent(core.ActionUp, 10, 15)
	handler.OnEvent(cb.Node(), up)

	if !received {
		t.Error("onChanged should have been called with true")
	}
	if !cb.IsChecked() {
		t.Error("checkbox should be checked after click")
	}
}

func TestCheckBoxText(t *testing.T) {
	cb := NewCheckBox("Hello")
	if cb.GetText() != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", cb.GetText())
	}
	cb.SetText("World")
	if cb.GetText() != "World" {
		t.Errorf("expected text 'World', got '%s'", cb.GetText())
	}
}

func TestCheckBoxDoubleClick(t *testing.T) {
	cb := NewCheckBox("Toggle")
	handler := cb.Node().GetHandler()
	cb.Node().SetBounds(core.Rect{Width: 200, Height: 30})

	// First click: unchecked -> checked
	handler.OnEvent(cb.Node(), core.NewMotionEvent(core.ActionDown, 10, 15))
	handler.OnEvent(cb.Node(), core.NewMotionEvent(core.ActionUp, 10, 15))
	if !cb.IsChecked() {
		t.Error("should be checked after first click")
	}

	// Second click: checked -> unchecked
	handler.OnEvent(cb.Node(), core.NewMotionEvent(core.ActionDown, 10, 15))
	handler.OnEvent(cb.Node(), core.NewMotionEvent(core.ActionUp, 10, 15))
	if cb.IsChecked() {
		t.Error("should be unchecked after second click")
	}
}
