package widget

import (
	"testing"
)

func TestNewTabLayout(t *testing.T) {
	tl := NewTabLayout()
	if tl.GetTabCount() != 0 {
		t.Errorf("expected 0 tabs, got %d", tl.GetTabCount())
	}
	if tl.GetSelectedTab() != 0 {
		t.Errorf("expected selected tab 0, got %d", tl.GetSelectedTab())
	}
	if tl.Node().Tag() != "TabLayout" {
		t.Errorf("expected tag 'TabLayout', got %q", tl.Node().Tag())
	}
}

func TestTabLayoutAddRemove(t *testing.T) {
	tl := NewTabLayout()
	tl.AddTab(Tab{Text: "Home"})
	tl.AddTab(Tab{Text: "Settings"})
	tl.AddTab(Tab{Text: "About"})

	if tl.GetTabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tl.GetTabCount())
	}

	tl.RemoveTabAt(1) // remove "Settings"
	if tl.GetTabCount() != 2 {
		t.Errorf("expected 2 tabs after remove, got %d", tl.GetTabCount())
	}
	if tl.tabs[1].Text != "About" {
		t.Errorf("expected second tab 'About', got %q", tl.tabs[1].Text)
	}
}

func TestTabLayoutSelection(t *testing.T) {
	tl := NewTabLayout()
	tl.AddTab(Tab{Text: "Tab1"})
	tl.AddTab(Tab{Text: "Tab2"})
	tl.AddTab(Tab{Text: "Tab3"})

	selectedIndex := -1
	tl.SetOnTabSelectedListener(func(index int) {
		selectedIndex = index
	})

	tl.SetSelectedTab(2)
	if tl.GetSelectedTab() != 2 {
		t.Errorf("expected selected tab 2, got %d", tl.GetSelectedTab())
	}
	if selectedIndex != 2 {
		t.Errorf("expected listener called with 2, got %d", selectedIndex)
	}

	// Same tab — no callback
	selectedIndex = -1
	tl.SetSelectedTab(2)
	if selectedIndex != -1 {
		t.Error("expected listener NOT called for same tab")
	}

	// Out of range — no change
	tl.SetSelectedTab(10)
	if tl.GetSelectedTab() != 2 {
		t.Errorf("expected selected tab still 2, got %d", tl.GetSelectedTab())
	}
}

func TestTabLayoutRemoveAdjustsSelection(t *testing.T) {
	tl := NewTabLayout()
	tl.AddTab(Tab{Text: "A"})
	tl.AddTab(Tab{Text: "B"})
	tl.SetSelectedTab(1)

	tl.RemoveTabAt(1) // remove selected tab
	if tl.GetSelectedTab() != 0 {
		t.Errorf("expected selected tab adjusted to 0, got %d", tl.GetSelectedTab())
	}
}
