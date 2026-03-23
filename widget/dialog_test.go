package widget

import (
	"testing"

	"gowui/core"
)

func TestAlertDialogBuilder(t *testing.T) {
	d := NewAlertDialogBuilder().
		SetTitle("Confirm").
		SetMessage("Are you sure?").
		SetPositiveButton("OK", nil).
		SetNegativeButton("Cancel", nil).
		Build()

	if d.GetTitle() != "Confirm" {
		t.Errorf("expected title 'Confirm', got %q", d.GetTitle())
	}
	if d.GetMessage() != "Are you sure?" {
		t.Errorf("expected message 'Are you sure?', got %q", d.GetMessage())
	}
	if d.positiveText != "OK" {
		t.Errorf("expected positive 'OK', got %q", d.positiveText)
	}
	if d.negativeText != "Cancel" {
		t.Errorf("expected negative 'Cancel', got %q", d.negativeText)
	}
	if d.Node().Tag() != "Dialog" {
		t.Errorf("expected tag 'Dialog', got %q", d.Node().Tag())
	}
}

func TestDialogShowDismiss(t *testing.T) {
	root := core.NewNode("Root")

	dismissed := false
	d := NewAlertDialogBuilder().
		SetTitle("Test").
		SetMessage("Hello").
		SetPositiveButton("OK", nil).
		SetOnDismissListener(func() { dismissed = true }).
		Build()

	d.ShowInNode(root)
	if !d.IsShowing() {
		t.Error("expected dialog to be showing")
	}
	if len(root.Children()) != 1 { // overlay
		t.Errorf("expected 1 root child (overlay), got %d", len(root.Children()))
	}

	d.Dismiss()
	if d.IsShowing() {
		t.Error("expected dialog not showing after dismiss")
	}
	if !dismissed {
		t.Error("expected dismiss listener to be called")
	}
	if len(root.Children()) != 0 {
		t.Errorf("expected 0 root children after dismiss, got %d", len(root.Children()))
	}
}

func TestDialogDoubleShow(t *testing.T) {
	root := core.NewNode("Root")
	d := NewAlertDialogBuilder().
		SetTitle("Test").
		SetPositiveButton("OK", nil).
		Build()

	d.ShowInNode(root)
	d.ShowInNode(root) // should not add a second overlay
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 overlay on double show, got %d", len(root.Children()))
	}
	d.Dismiss()
}

func TestDialogDoubleDismiss(t *testing.T) {
	root := core.NewNode("Root")
	dismissCount := 0
	d := NewAlertDialogBuilder().
		SetTitle("Test").
		SetOnDismissListener(func() { dismissCount++ }).
		Build()

	d.ShowInNode(root)
	d.Dismiss()
	d.Dismiss() // should not call listener again
	if dismissCount != 1 {
		t.Errorf("expected dismiss called once, got %d", dismissCount)
	}
}

func TestDialogButtonTexts(t *testing.T) {
	d := NewAlertDialogBuilder().
		SetPositiveButton("OK", nil).
		SetNegativeButton("Cancel", nil).
		SetNeutralButton("Help", nil).
		Build()

	btns := d.buttonTexts()
	if len(btns) != 3 {
		t.Fatalf("expected 3 buttons, got %d", len(btns))
	}
	// Order: neutral, negative, positive
	if btns[0].text != "Help" {
		t.Errorf("expected first button 'Help', got %q", btns[0].text)
	}
	if btns[1].text != "Cancel" {
		t.Errorf("expected second button 'Cancel', got %q", btns[1].text)
	}
	if btns[2].text != "OK" {
		t.Errorf("expected third button 'OK', got %q", btns[2].text)
	}
}

func TestDialogPositiveClick(t *testing.T) {
	root := core.NewNode("Root")
	clicked := false
	d := NewAlertDialogBuilder().
		SetTitle("Test").
		SetPositiveButton("OK", func() { clicked = true }).
		Build()

	d.ShowInNode(root)

	// Directly invoke the click handler to verify wiring
	if d.positiveClick != nil {
		d.positiveClick()
	}
	if !clicked {
		t.Error("expected positive click handler to be called")
	}
	d.Dismiss()
}

func TestDialogCancelable(t *testing.T) {
	d := NewAlertDialogBuilder().
		SetCancelable(false).
		Build()

	if d.cancelable {
		t.Error("expected cancelable to be false")
	}
}
