package widget

import (
	"testing"
)

func TestNewSpinner(t *testing.T) {
	sp := NewSpinner([]string{"Apple", "Banana", "Cherry"})
	if sp.Node().Tag() != "Spinner" {
		t.Errorf("expected tag 'Spinner', got %q", sp.Node().Tag())
	}
	if sp.GetSelectedIndex() != 0 {
		t.Errorf("expected selected index 0, got %d", sp.GetSelectedIndex())
	}
	if sp.GetSelectedItem() != "Apple" {
		t.Errorf("expected 'Apple', got %q", sp.GetSelectedItem())
	}
}

func TestSpinnerEmpty(t *testing.T) {
	sp := NewSpinner(nil)
	if sp.GetSelectedIndex() != -1 {
		t.Errorf("expected -1 for empty, got %d", sp.GetSelectedIndex())
	}
	if sp.GetSelectedItem() != "" {
		t.Errorf("expected empty string, got %q", sp.GetSelectedItem())
	}
}

func TestSpinnerSetSelectedIndex(t *testing.T) {
	sp := NewSpinner([]string{"A", "B", "C"})
	sp.SetSelectedIndex(2)
	if sp.GetSelectedIndex() != 2 {
		t.Errorf("expected 2, got %d", sp.GetSelectedIndex())
	}
	if sp.GetSelectedItem() != "C" {
		t.Errorf("expected 'C', got %q", sp.GetSelectedItem())
	}

	// Out of range
	sp.SetSelectedIndex(10)
	if sp.GetSelectedIndex() != 2 {
		t.Errorf("expected still 2, got %d", sp.GetSelectedIndex())
	}
}

func TestSpinnerSetItems(t *testing.T) {
	sp := NewSpinner([]string{"A", "B"})
	sp.SetSelectedIndex(1)

	sp.SetItems([]string{"X"})
	if sp.GetSelectedIndex() != 0 {
		t.Errorf("expected adjusted to 0, got %d", sp.GetSelectedIndex())
	}
	if sp.GetSelectedItem() != "X" {
		t.Errorf("expected 'X', got %q", sp.GetSelectedItem())
	}
}

func TestSpinnerOnSelected(t *testing.T) {
	sp := NewSpinner([]string{"A", "B", "C"})
	selectedIdx := -1
	selectedItem := ""
	sp.SetOnItemSelectedListener(func(idx int, item string) {
		selectedIdx = idx
		selectedItem = item
	})

	// Simulate selection by directly invoking callback
	sp.onSelected(1, "B")
	if selectedIdx != 1 || selectedItem != "B" {
		t.Errorf("expected (1, B), got (%d, %q)", selectedIdx, selectedItem)
	}
}

func TestSpinnerGetItems(t *testing.T) {
	items := []string{"Red", "Green", "Blue"}
	sp := NewSpinner(items)
	got := sp.GetItems()
	if len(got) != 3 {
		t.Errorf("expected 3 items, got %d", len(got))
	}
}
