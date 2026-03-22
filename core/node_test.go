package core

import "testing"

func TestNodeAddChild(t *testing.T) {
	parent := NewNode("LinearLayout")
	child1 := NewNode("TextView")
	child2 := NewNode("Button")
	parent.AddChild(child1)
	parent.AddChild(child2)
	if len(parent.Children()) != 2 {
		t.Fatalf("expected 2 children")
	}
	if child1.Parent() != parent {
		t.Error("child1 parent mismatch")
	}
	if child2.Parent() != parent {
		t.Error("child2 parent mismatch")
	}
}

func TestNodeRemoveChild(t *testing.T) {
	parent := NewNode("LinearLayout")
	child := NewNode("TextView")
	parent.AddChild(child)
	parent.RemoveChild(child)
	if len(parent.Children()) != 0 {
		t.Error("expected 0 children")
	}
	if child.Parent() != nil {
		t.Error("parent should be nil after remove")
	}
}

func TestNodeFindById(t *testing.T) {
	root := NewNode("LinearLayout")
	child := NewNode("TextView")
	child.SetId("title")
	grandchild := NewNode("Button")
	grandchild.SetId("btn_ok")
	root.AddChild(child)
	child.AddChild(grandchild)

	found := root.FindNodeById("title")
	if found != child {
		t.Error("should find child by id")
	}
	found2 := root.FindNodeById("btn_ok")
	if found2 != grandchild {
		t.Error("should find grandchild by id")
	}
	notFound := root.FindNodeById("nonexistent")
	if notFound != nil {
		t.Error("should return nil for missing id")
	}
}

func TestNodeDirtyBubble(t *testing.T) {
	root := NewNode("Root")
	child := NewNode("Child")
	grandchild := NewNode("Leaf")
	root.AddChild(child)
	child.AddChild(grandchild)

	root.ClearDirty()
	child.ClearDirty()
	grandchild.ClearDirty()

	grandchild.MarkDirty()
	if !grandchild.IsDirty() {
		t.Error("grandchild should be dirty")
	}
	if !child.IsChildDirty() {
		t.Error("child should have childDirty")
	}
	if !root.IsChildDirty() {
		t.Error("root should have childDirty")
	}
}

func TestNodeFindViewById(t *testing.T) {
	root := NewNode("Root")
	child := NewNode("Button")
	child.SetId("btn")
	root.AddChild(child)

	// No view associated — should return nil
	if root.FindViewById("btn") != nil {
		t.Error("should return nil when no view set")
	}

	// Associate a mock view
	mockView := &mockViewImpl{node: child}
	child.SetView(mockView)
	found := root.FindViewById("btn")
	if found != mockView {
		t.Error("should find the associated view")
	}
}

// mockViewImpl for testing
type mockViewImpl struct {
	node *Node
}

func (m *mockViewImpl) Node() *Node               { return m.node }
func (m *mockViewImpl) SetId(id string)            { m.node.SetId(id) }
func (m *mockViewImpl) GetId() string              { return m.node.GetId() }
func (m *mockViewImpl) SetVisibility(v Visibility) { m.node.SetVisibility(v) }
func (m *mockViewImpl) GetVisibility() Visibility  { return m.node.GetVisibility() }
func (m *mockViewImpl) SetEnabled(b bool)          { m.node.SetEnabled(b) }
func (m *mockViewImpl) IsEnabled() bool            { return m.node.IsEnabled() }
