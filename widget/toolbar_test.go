package widget

import (
	"testing"
)

func TestNewToolbar(t *testing.T) {
	tb := NewToolbar("My App")
	if tb.GetTitle() != "My App" {
		t.Errorf("expected title 'My App', got %q", tb.GetTitle())
	}
	if tb.GetSubtitle() != "" {
		t.Errorf("expected empty subtitle, got %q", tb.GetSubtitle())
	}
	if tb.Node() == nil {
		t.Fatal("expected non-nil node")
	}
	if tb.Node().Tag() != "Toolbar" {
		t.Errorf("expected tag 'Toolbar', got %q", tb.Node().Tag())
	}
}

func TestToolbarSetTitle(t *testing.T) {
	tb := NewToolbar("Old Title")
	tb.SetTitle("New Title")
	if tb.GetTitle() != "New Title" {
		t.Errorf("expected 'New Title', got %q", tb.GetTitle())
	}
}

func TestToolbarSetSubtitle(t *testing.T) {
	tb := NewToolbar("Title")
	tb.SetSubtitle("Subtitle")
	if tb.GetSubtitle() != "Subtitle" {
		t.Errorf("expected 'Subtitle', got %q", tb.GetSubtitle())
	}
}

func TestToolbarActions(t *testing.T) {
	tb := NewToolbar("App")
	if tb.GetActionCount() != 0 {
		t.Errorf("expected 0 actions, got %d", tb.GetActionCount())
	}

	clicked := false
	tb.AddAction(ActionItem{
		ID:      "search",
		Title:   "Search",
		OnClick: func() { clicked = true },
	})
	tb.AddAction(ActionItem{
		ID:    "more",
		Title: "More",
	})

	if tb.GetActionCount() != 2 {
		t.Errorf("expected 2 actions, got %d", tb.GetActionCount())
	}

	// Simulate action click
	tb.actions[0].OnClick()
	if !clicked {
		t.Error("expected action click handler to be called")
	}

	tb.ClearActions()
	if tb.GetActionCount() != 0 {
		t.Errorf("expected 0 actions after clear, got %d", tb.GetActionCount())
	}
}

func TestToolbarNavigation(t *testing.T) {
	tb := NewToolbar("App")
	navClicked := false
	tb.SetNavigationOnClickListener(func() { navClicked = true })

	if tb.navText == "" {
		t.Error("expected default nav text to be set")
	}

	tb.SetNavigationText("☰")
	if tb.navText != "☰" {
		t.Errorf("expected nav text '☰', got %q", tb.navText)
	}

	tb.navOnClick()
	if !navClicked {
		t.Error("expected nav click handler to be called")
	}
}
