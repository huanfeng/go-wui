package widget

import (
	"gowui/core"
	"testing"
)

func TestButtonClickListener(t *testing.T) {
	btn := NewButton("OK", nil)
	clicked := false
	btn.SetOnClickListener(func(v core.View) { clicked = true })

	btn.Node().SetBounds(core.Rect{X: 0, Y: 0, Width: 100, Height: 48})
	handler := btn.Node().GetHandler()

	down := core.NewMotionEvent(core.ActionDown, 50, 24)
	handler.OnEvent(btn.Node(), down)

	up := core.NewMotionEvent(core.ActionUp, 50, 24)
	handler.OnEvent(btn.Node(), up)

	if !clicked {
		t.Error("click listener not called after down+up")
	}
}

func TestButtonStateTransitions(t *testing.T) {
	btn := NewButton("Test", nil)
	handler := btn.Node().GetHandler()

	if btn.State() != ButtonStateNormal {
		t.Error("initial state should be Normal")
	}

	hover := core.NewMotionEvent(core.ActionHoverEnter, 0, 0)
	handler.OnEvent(btn.Node(), hover)
	if btn.State() != ButtonStateHovered {
		t.Error("should be Hovered after HoverEnter")
	}

	down := core.NewMotionEvent(core.ActionDown, 0, 0)
	handler.OnEvent(btn.Node(), down)
	if btn.State() != ButtonStatePressed {
		t.Error("should be Pressed after Down")
	}

	up := core.NewMotionEvent(core.ActionUp, 0, 0)
	handler.OnEvent(btn.Node(), up)
	if btn.State() != ButtonStateNormal {
		t.Error("should be Normal after Up")
	}
}

func TestButtonTextUpdate(t *testing.T) {
	btn := NewButton("OK", nil)
	if btn.GetText() != "OK" {
		t.Error("initial text")
	}
	btn.SetText("Cancel")
	if btn.GetText() != "Cancel" {
		t.Error("updated text")
	}
}
