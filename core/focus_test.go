package core

import "testing"

// makeHandler 创建一个简单的事件处理器（返回 false 不消耗事件）。
func makeHandler() Handler {
	return &DefaultHandler{}
}

func TestFocusManagerRequestFocus(t *testing.T) {
	fm := NewFocusManager()
	n := NewNode("test")

	if fm.Current() != nil {
		t.Error("初始焦点应为 nil")
	}

	fm.RequestFocus(n)
	if fm.Current() != n {
		t.Error("RequestFocus 后焦点应为该节点")
	}

	fm.ClearFocus()
	if fm.Current() != nil {
		t.Error("ClearFocus 后焦点应为 nil")
	}
}

func TestFocusManagerSetFocused(t *testing.T) {
	fm := NewFocusManager()
	n := NewNode("test")

	fm.SetFocused(n)
	if fm.Current() != n {
		t.Error("SetFocused 后焦点应为该节点")
	}
}

func TestMoveFocusForward(t *testing.T) {
	fm := NewFocusManager()

	// 构建根节点和三个可聚焦子节点
	root := NewNode("root")
	n1 := NewNode("n1")
	n2 := NewNode("n2")
	n3 := NewNode("n3")
	n1.SetHandler(makeHandler())
	n2.SetHandler(makeHandler())
	n3.SetHandler(makeHandler())
	root.AddChild(n1)
	root.AddChild(n2)
	root.AddChild(n3)

	// 初始无焦点 → Tab 应聚焦到第一个节点
	fm.MoveFocus(root, true)
	if fm.Current() != n1 {
		t.Errorf("第一次 Tab 应聚焦 n1，实际: %v", fm.Current())
	}

	// n1 → Tab → n2
	fm.MoveFocus(root, true)
	if fm.Current() != n2 {
		t.Errorf("第二次 Tab 应聚焦 n2，实际: %v", fm.Current())
	}

	// n2 → Tab → n3
	fm.MoveFocus(root, true)
	if fm.Current() != n3 {
		t.Errorf("第三次 Tab 应聚焦 n3，实际: %v", fm.Current())
	}

	// n3 → Tab → n1（循环）
	fm.MoveFocus(root, true)
	if fm.Current() != n1 {
		t.Errorf("循环 Tab 应回到 n1，实际: %v", fm.Current())
	}
}

func TestMoveFocusBackward(t *testing.T) {
	fm := NewFocusManager()

	root := NewNode("root")
	n1 := NewNode("n1")
	n2 := NewNode("n2")
	n3 := NewNode("n3")
	n1.SetHandler(makeHandler())
	n2.SetHandler(makeHandler())
	n3.SetHandler(makeHandler())
	root.AddChild(n1)
	root.AddChild(n2)
	root.AddChild(n3)

	// 初始无焦点 → Shift+Tab 应聚焦到最后一个节点
	fm.MoveFocus(root, false)
	if fm.Current() != n3 {
		t.Errorf("初始 Shift+Tab 应聚焦 n3，实际: %v", fm.Current())
	}

	// n3 → Shift+Tab → n2
	fm.MoveFocus(root, false)
	if fm.Current() != n2 {
		t.Errorf("Shift+Tab 应聚焦 n2，实际: %v", fm.Current())
	}

	// n2 → Shift+Tab → n1
	fm.MoveFocus(root, false)
	if fm.Current() != n1 {
		t.Errorf("Shift+Tab 应聚焦 n1，实际: %v", fm.Current())
	}

	// n1 → Shift+Tab → n3（循环）
	fm.MoveFocus(root, false)
	if fm.Current() != n3 {
		t.Errorf("循环 Shift+Tab 应回到 n3，实际: %v", fm.Current())
	}
}

func TestMoveFocusSkipsDisabled(t *testing.T) {
	fm := NewFocusManager()

	root := NewNode("root")
	n1 := NewNode("n1")
	n2 := NewNode("n2") // 禁用
	n3 := NewNode("n3")
	n1.SetHandler(makeHandler())
	n2.SetHandler(makeHandler())
	n2.SetEnabled(false) // 禁用节点不可聚焦
	n3.SetHandler(makeHandler())
	root.AddChild(n1)
	root.AddChild(n2)
	root.AddChild(n3)

	fm.MoveFocus(root, true) // → n1
	fm.MoveFocus(root, true) // n1 → n3（跳过 n2）
	if fm.Current() != n3 {
		t.Errorf("应跳过禁用节点聚焦 n3，实际: %v", fm.Current())
	}
}

func TestMoveFocusSkipsInvisible(t *testing.T) {
	fm := NewFocusManager()

	root := NewNode("root")
	n1 := NewNode("n1")
	n2 := NewNode("n2") // 不可见
	n3 := NewNode("n3")
	n1.SetHandler(makeHandler())
	n2.SetHandler(makeHandler())
	n2.SetVisibility(Invisible) // 不可见节点不可聚焦
	n3.SetHandler(makeHandler())
	root.AddChild(n1)
	root.AddChild(n2)
	root.AddChild(n3)

	fm.MoveFocus(root, true) // → n1
	fm.MoveFocus(root, true) // n1 → n3（跳过不可见 n2）
	if fm.Current() != n3 {
		t.Errorf("应跳过不可见节点聚焦 n3，实际: %v", fm.Current())
	}
}

func TestMoveFocusEmpty(t *testing.T) {
	fm := NewFocusManager()
	root := NewNode("root")

	// 无可聚焦节点，不应 panic
	fm.MoveFocus(root, true)
	if fm.Current() != nil {
		t.Error("无可聚焦节点时焦点应保持 nil")
	}
}
