package widget

import (
	"testing"

	"github.com/huanfeng/wind-ui/core"
)

func TestNewMenu(t *testing.T) {
	m := NewMenu()
	if m.GetItemCount() != 0 {
		t.Errorf("expected 0 items, got %d", m.GetItemCount())
	}
}

func TestMenuAddItem(t *testing.T) {
	m := NewMenu()
	clicked := false
	m.AddItem("cut", "Cut", func() { clicked = true })
	m.AddItem("copy", "Copy", nil)
	m.AddItem("paste", "Paste", nil)

	if m.GetItemCount() != 3 {
		t.Errorf("expected 3 items, got %d", m.GetItemCount())
	}

	items := m.GetItems()
	if items[0].Title != "Cut" {
		t.Errorf("expected 'Cut', got %q", items[0].Title)
	}
	if !items[0].Enabled {
		t.Error("expected item to be enabled")
	}

	items[0].OnClick()
	if !clicked {
		t.Error("expected click handler to be called")
	}
}

func TestMenuClear(t *testing.T) {
	m := NewMenu()
	m.AddItem("a", "A", nil)
	m.AddItem("b", "B", nil)
	m.Clear()
	if m.GetItemCount() != 0 {
		t.Errorf("expected 0 items after clear, got %d", m.GetItemCount())
	}
}

func TestMenuAdd(t *testing.T) {
	m := NewMenu()
	m.Add(MenuItem{ID: "test", Title: "Test", Enabled: true})
	if m.GetItemCount() != 1 {
		t.Errorf("expected 1 item, got %d", m.GetItemCount())
	}
}

func TestNewPopupMenu(t *testing.T) {
	m := NewMenu()
	m.AddItem("a", "Action A", nil)
	pm := NewPopupMenu(m)

	if pm.Node().Tag() != "PopupMenu" {
		t.Errorf("expected tag 'PopupMenu', got %q", pm.Node().Tag())
	}
	if pm.IsShowing() {
		t.Error("expected popup not showing initially")
	}
}

func TestPopupMenuShowDismiss(t *testing.T) {
	m := NewMenu()
	m.AddItem("a", "Action A", nil)
	m.AddItem("b", "Action B", nil)
	pm := NewPopupMenu(m)

	// Create an anchor node tree
	root := core.NewNode("Root")
	anchor := core.NewNode("Anchor")
	root.AddChild(anchor)

	dismissed := false
	pm.SetOnDismissListener(func() { dismissed = true })

	pm.ShowAtPosition(anchor, 100, 100)
	if !pm.IsShowing() {
		t.Error("expected popup to be showing")
	}
	// Overlay should be added to root
	if len(root.Children()) != 2 { // anchor + overlay
		t.Errorf("expected 2 root children, got %d", len(root.Children()))
	}

	pm.Dismiss()
	if pm.IsShowing() {
		t.Error("expected popup not showing after dismiss")
	}
	if !dismissed {
		t.Error("expected dismiss listener to be called")
	}
	// Overlay should be removed
	if len(root.Children()) != 1 {
		t.Errorf("expected 1 root child after dismiss, got %d", len(root.Children()))
	}
}

func TestPopupMenuDimensions(t *testing.T) {
	m := NewMenu()
	m.AddItem("a", "Short", nil)
	m.AddItem("b", "A Longer Menu Item", nil)
	pm := NewPopupMenu(m)

	w := pm.menuWidth()
	h := pm.menuHeight()

	if w < menuMinWidth {
		t.Errorf("expected width >= %f, got %f", menuMinWidth, w)
	}
	if w > menuMaxWidth {
		t.Errorf("expected width <= %f, got %f", menuMaxWidth, w)
	}
	if h != float64(2)*menuItemHeight {
		t.Errorf("expected height %f, got %f", float64(2)*menuItemHeight, h)
	}
}
