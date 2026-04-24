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

// TestScaleNodeDPI 验证首次 DPI 缩放的正确性。
func TestScaleNodeDPI(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{
		FontSize:     14,
		CornerRadius: 4,
		BorderWidth:  1,
		Width:        Dimension{Unit: DimensionDp, Value: 100},
		Height:       Dimension{Unit: DimensionDp, Value: 50},
	})
	node.SetPadding(Insets{Left: 8, Top: 4, Right: 8, Bottom: 4})
	node.SetMargin(Insets{Left: 2, Top: 2, Right: 2, Bottom: 2})

	ScaleNodeDPI(node, 1.5)

	s := node.GetStyle()
	if s.FontSize != 21 {
		t.Errorf("FontSize: got %.2f, want 21", s.FontSize)
	}
	if s.CornerRadius != 6 {
		t.Errorf("CornerRadius: got %.2f, want 6", s.CornerRadius)
	}
	if s.Width.Value != 150 {
		t.Errorf("Width: got %.2f, want 150", s.Width.Value)
	}
	p := node.Padding()
	if p.Left != 12 {
		t.Errorf("Padding.Left: got %.2f, want 12", p.Left)
	}
	// dpiScale 应已存储到节点
	if scale, ok := node.GetData("dpiScale").(float64); !ok || scale != 1.5 {
		t.Errorf("dpiScale: got %v, want 1.5", node.GetData("dpiScale"))
	}
}

// TestScaleNodeDPI_NoDoubleScale 验证重复调用 ScaleNodeDPI 不会重复缩放。
func TestScaleNodeDPI_NoDoubleScale(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{FontSize: 14})

	ScaleNodeDPI(node, 1.5)
	ScaleNodeDPI(node, 1.5) // 第二次应被跳过

	s := node.GetStyle()
	if s.FontSize != 21 {
		t.Errorf("FontSize after double scale: got %.2f, want 21", s.FontSize)
	}
}

// TestScaleNodeDPI_ScaleOne 验证 scale=1.0 时不改变值但标记节点。
func TestScaleNodeDPI_ScaleOne(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{FontSize: 14})

	ScaleNodeDPI(node, 1.0)

	s := node.GetStyle()
	if s.FontSize != 14 {
		t.Errorf("FontSize with scale=1.0: got %.2f, want 14", s.FontSize)
	}
	if node.GetData("dpiScale").(float64) != 1.0 {
		t.Errorf("dpiScale should be 1.0")
	}
}

// TestRescaleNodeDPI_FromOriginal 验证 RescaleNodeDPI 从原始值重新计算，无累积误差。
func TestRescaleNodeDPI_FromOriginal(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{
		FontSize: 14,
		Width:    Dimension{Unit: DimensionDp, Value: 100},
	})
	node.SetPadding(Insets{Left: 8, Top: 4, Right: 8, Bottom: 4})

	// 首次缩放到 1.5x
	ScaleNodeDPI(node, 1.5)

	// 重新缩放到 2.0x（应从原始值 14 * 2.0 计算，而非 21 * (2/1.5)）
	RescaleNodeDPI(node, 2.0)

	s := node.GetStyle()
	if s.FontSize != 28 {
		t.Errorf("FontSize after rescale to 2.0x: got %.4f, want 28", s.FontSize)
	}
	if s.Width.Value != 200 {
		t.Errorf("Width after rescale to 2.0x: got %.4f, want 200", s.Width.Value)
	}
	p := node.Padding()
	if p.Left != 16 {
		t.Errorf("Padding.Left after rescale to 2.0x: got %.4f, want 16", p.Left)
	}
	if scale := node.GetData("dpiScale").(float64); scale != 2.0 {
		t.Errorf("dpiScale: got %v, want 2.0", scale)
	}
}

// TestRescaleNodeDPI_NoCumulativeError 验证多次重缩放不产生浮点累积误差。
func TestRescaleNodeDPI_NoCumulativeError(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{FontSize: 10})

	ScaleNodeDPI(node, 1.0)
	// 模拟多次 DPI 变化
	RescaleNodeDPI(node, 1.25)
	RescaleNodeDPI(node, 1.5)
	RescaleNodeDPI(node, 1.75)
	RescaleNodeDPI(node, 2.0)

	s := node.GetStyle()
	// 每次都从原始值 10 计算，最终应为 10 * 2.0 = 20，无浮点误差
	if s.FontSize != 20 {
		t.Errorf("FontSize after multiple rescales: got %.6f, want 20", s.FontSize)
	}
}

// TestRescaleNodeDPI_Recursive 验证 RescaleNodeDPI 递归处理子节点。
func TestRescaleNodeDPI_Recursive(t *testing.T) {
	parent := NewNode("LinearLayout")
	child := NewNode("TextView")
	child.SetStyle(&Style{FontSize: 12})

	parent.AddChild(child)
	ScaleNodeDPI(parent, 1.5)

	// 验证子节点已被首次缩放
	if child.GetStyle().FontSize != 18 {
		t.Errorf("child FontSize after ScaleNodeDPI: got %.2f, want 18", child.GetStyle().FontSize)
	}

	// 重新缩放整棵树
	RescaleNodeDPI(parent, 2.0)

	if child.GetStyle().FontSize != 24 {
		t.Errorf("child FontSize after RescaleNodeDPI to 2.0x: got %.2f, want 24", child.GetStyle().FontSize)
	}
}

// TestRescaleNodeDPI_SameScale 验证相同 scale 不重复计算。
func TestRescaleNodeDPI_SameScale(t *testing.T) {
	node := NewNode("TextView")
	node.SetStyle(&Style{FontSize: 14})
	ScaleNodeDPI(node, 1.5)

	// 以相同比例重缩放，值不应改变
	RescaleNodeDPI(node, 1.5)

	if node.GetStyle().FontSize != 21 {
		t.Errorf("FontSize should remain 21, got %.2f", node.GetStyle().FontSize)
	}
}
