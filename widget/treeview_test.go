package widget

import (
	"testing"
)

func TestNewTreeView(t *testing.T) {
	tv := NewTreeView()
	if tv.Node().Tag() != "TreeView" {
		t.Errorf("expected tag 'TreeView', got %q", tv.Node().Tag())
	}
	if tv.GetFlatCount() != 0 {
		t.Errorf("expected 0 flat entries, got %d", tv.GetFlatCount())
	}
}

func TestTreeViewSetRoots(t *testing.T) {
	tv := NewTreeView()
	root1 := &TreeNode{Text: "Root 1"}
	root2 := &TreeNode{Text: "Root 2"}
	tv.SetRoots([]*TreeNode{root1, root2})

	if tv.GetFlatCount() != 2 {
		t.Errorf("expected 2 flat entries, got %d", tv.GetFlatCount())
	}
}

func TestTreeViewExpandCollapse(t *testing.T) {
	tv := NewTreeView()
	root := &TreeNode{Text: "Root", Expanded: false}
	root.AddChild(&TreeNode{Text: "Child 1"})
	root.AddChild(&TreeNode{Text: "Child 2"})
	tv.SetRoots([]*TreeNode{root})

	// Collapsed: only root visible
	if tv.GetFlatCount() != 1 {
		t.Errorf("expected 1 flat entry when collapsed, got %d", tv.GetFlatCount())
	}

	// Expand
	root.Expanded = true
	tv.rebuildFlat()
	if tv.GetFlatCount() != 3 {
		t.Errorf("expected 3 flat entries when expanded, got %d", tv.GetFlatCount())
	}

	// Collapse
	root.Expanded = false
	tv.rebuildFlat()
	if tv.GetFlatCount() != 1 {
		t.Errorf("expected 1 flat entry after collapse, got %d", tv.GetFlatCount())
	}
}

func TestTreeViewNestedExpand(t *testing.T) {
	tv := NewTreeView()
	root := &TreeNode{Text: "Root", Expanded: true}
	child := &TreeNode{Text: "Child", Expanded: true}
	child.AddChild(&TreeNode{Text: "Grandchild"})
	root.AddChild(child)
	tv.SetRoots([]*TreeNode{root})

	// Root + Child + Grandchild = 3
	if tv.GetFlatCount() != 3 {
		t.Errorf("expected 3 flat entries, got %d", tv.GetFlatCount())
	}

	// Check depths
	if tv.flat[0].depth != 0 {
		t.Errorf("expected root depth 0, got %d", tv.flat[0].depth)
	}
	if tv.flat[1].depth != 1 {
		t.Errorf("expected child depth 1, got %d", tv.flat[1].depth)
	}
	if tv.flat[2].depth != 2 {
		t.Errorf("expected grandchild depth 2, got %d", tv.flat[2].depth)
	}
}

func TestTreeNodeIsLeaf(t *testing.T) {
	leaf := &TreeNode{Text: "Leaf"}
	if !leaf.IsLeaf() {
		t.Error("expected leaf")
	}

	parent := &TreeNode{Text: "Parent"}
	parent.AddChild(&TreeNode{Text: "Child"})
	if parent.IsLeaf() {
		t.Error("expected non-leaf")
	}
}

func TestTreeViewAddRoot(t *testing.T) {
	tv := NewTreeView()
	tv.AddRoot(&TreeNode{Text: "A"})
	tv.AddRoot(&TreeNode{Text: "B"})

	if len(tv.GetRoots()) != 2 {
		t.Errorf("expected 2 roots, got %d", len(tv.GetRoots()))
	}
	if tv.GetFlatCount() != 2 {
		t.Errorf("expected 2 flat entries, got %d", tv.GetFlatCount())
	}
}

func TestTreeViewOnSelected(t *testing.T) {
	tv := NewTreeView()
	root := &TreeNode{Text: "Root"}
	tv.SetRoots([]*TreeNode{root})

	var selectedNode *TreeNode
	tv.SetOnNodeSelectedListener(func(n *TreeNode) {
		selectedNode = n
	})

	tv.selectedIdx = 0
	tv.onSelected(root)
	if selectedNode != root {
		t.Error("expected root to be selected")
	}
}

func TestTreeViewGetSelectedNode(t *testing.T) {
	tv := NewTreeView()
	tv.SetRoots([]*TreeNode{
		{Text: "A"},
		{Text: "B"},
	})

	if tv.GetSelectedNode() != nil {
		t.Error("expected nil when nothing selected")
	}

	tv.selectedIdx = 1
	sel := tv.GetSelectedNode()
	if sel == nil || sel.Text != "B" {
		t.Errorf("expected 'B', got %v", sel)
	}
}
