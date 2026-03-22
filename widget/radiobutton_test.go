package widget

import (
	"gowui/core"
	"testing"
)

func TestRadioButtonDefault(t *testing.T) {
	rb := NewRadioButton("Option 1")
	if rb.IsSelected() {
		t.Error("should start unselected")
	}
	if rb.GetText() != "Option 1" {
		t.Errorf("expected text 'Option 1', got '%s'", rb.GetText())
	}
}

func TestRadioButtonSetSelected(t *testing.T) {
	rb := NewRadioButton("Option")
	rb.SetSelected(true)
	if !rb.IsSelected() {
		t.Error("should be selected")
	}
	rb.SetSelected(false)
	if rb.IsSelected() {
		t.Error("should be unselected")
	}
}

func TestRadioButtonOnChanged(t *testing.T) {
	rb := NewRadioButton("Option")
	var received bool
	rb.SetOnSelectedChanged(func(selected bool) { received = selected })

	// Simulate click via handler
	handler := rb.Node().GetHandler()
	rb.Node().SetBounds(core.Rect{Width: 200, Height: 30})

	down := core.NewMotionEvent(core.ActionDown, 10, 15)
	handler.OnEvent(rb.Node(), down)

	up := core.NewMotionEvent(core.ActionUp, 10, 15)
	handler.OnEvent(rb.Node(), up)

	if !received {
		t.Error("onChanged should have been called with true")
	}
	if !rb.IsSelected() {
		t.Error("radio button should be selected after click")
	}
}

func TestRadioButtonText(t *testing.T) {
	rb := NewRadioButton("Hello")
	if rb.GetText() != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", rb.GetText())
	}
	rb.SetText("World")
	if rb.GetText() != "World" {
		t.Errorf("expected text 'World', got '%s'", rb.GetText())
	}
}

func TestRadioGroupSelection(t *testing.T) {
	rg := NewRadioGroup()
	rb1 := NewRadioButton("Option 1")
	rb2 := NewRadioButton("Option 2")
	rb3 := NewRadioButton("Option 3")
	rg.AddButton(rb1)
	rg.AddButton(rb2)
	rg.AddButton(rb3)

	if rg.GetSelectedIndex() != -1 {
		t.Error("should start with no selection")
	}

	rg.SetSelectedIndex(1)
	if !rb2.IsSelected() {
		t.Error("rb2 should be selected")
	}
	if rb1.IsSelected() {
		t.Error("rb1 should not be selected")
	}

	rg.SetSelectedIndex(0)
	if !rb1.IsSelected() {
		t.Error("rb1 should now be selected")
	}
	if rb2.IsSelected() {
		t.Error("rb2 should be deselected")
	}
}

func TestRadioGroupOnChanged(t *testing.T) {
	rg := NewRadioGroup()
	rb1 := NewRadioButton("A")
	rb2 := NewRadioButton("B")
	rg.AddButton(rb1)
	rg.AddButton(rb2)

	changedIdx := -1
	rg.SetOnChanged(func(idx int) { changedIdx = idx })
	rg.SetSelectedIndex(1)
	if changedIdx != 1 {
		t.Errorf("expected 1, got %d", changedIdx)
	}
}

func TestRadioGroupClickSelection(t *testing.T) {
	rg := NewRadioGroup()
	rb1 := NewRadioButton("A")
	rb2 := NewRadioButton("B")
	rg.AddButton(rb1)
	rg.AddButton(rb2)

	changedIdx := -1
	rg.SetOnChanged(func(idx int) { changedIdx = idx })

	// Simulate click on rb2 via its handler
	handler := rb2.Node().GetHandler()
	rb2.Node().SetBounds(core.Rect{Width: 200, Height: 30})

	handler.OnEvent(rb2.Node(), core.NewMotionEvent(core.ActionDown, 10, 15))
	handler.OnEvent(rb2.Node(), core.NewMotionEvent(core.ActionUp, 10, 15))

	if !rb2.IsSelected() {
		t.Error("rb2 should be selected after click")
	}
	if rb1.IsSelected() {
		t.Error("rb1 should be deselected")
	}
	if changedIdx != 1 {
		t.Errorf("expected onChanged index 1, got %d", changedIdx)
	}

	// Click on rb1
	handler1 := rb1.Node().GetHandler()
	rb1.Node().SetBounds(core.Rect{Width: 200, Height: 30})

	handler1.OnEvent(rb1.Node(), core.NewMotionEvent(core.ActionDown, 10, 15))
	handler1.OnEvent(rb1.Node(), core.NewMotionEvent(core.ActionUp, 10, 15))

	if !rb1.IsSelected() {
		t.Error("rb1 should be selected after click")
	}
	if rb2.IsSelected() {
		t.Error("rb2 should be deselected after rb1 clicked")
	}
	if changedIdx != 0 {
		t.Errorf("expected onChanged index 0, got %d", changedIdx)
	}
}

func TestRadioGroupSetSelectedIndexOutOfRange(t *testing.T) {
	rg := NewRadioGroup()
	rb1 := NewRadioButton("A")
	rg.AddButton(rb1)

	// Out of range should be a no-op
	rg.SetSelectedIndex(5)
	if rg.GetSelectedIndex() != -1 {
		t.Error("out-of-range index should not change selection")
	}

	rg.SetSelectedIndex(-2)
	if rg.GetSelectedIndex() != -1 {
		t.Error("negative index should not change selection")
	}
}
