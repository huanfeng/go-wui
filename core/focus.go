package core

// FocusManager tracks which node currently holds input focus.
type FocusManager struct {
	current *Node
}

// NewFocusManager creates a FocusManager with no focused node.
func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

// Current returns the currently focused node, or nil.
func (fm *FocusManager) Current() *Node { return fm.current }

// RequestFocus moves focus to the given node.
func (fm *FocusManager) RequestFocus(node *Node) {
	if fm.current == node {
		return
	}
	fm.current = node
}

// SetFocused 直接将焦点设置到指定节点（RequestFocus 的公开别名）。
func (fm *FocusManager) SetFocused(node *Node) {
	fm.RequestFocus(node)
}

// ClearFocus removes focus from any node.
func (fm *FocusManager) ClearFocus() {
	fm.current = nil
}

// collectFocusable 递归收集所有可聚焦节点（enabled、visible、有 handler）。
func collectFocusable(node *Node, result *[]*Node) {
	if node == nil {
		return
	}
	if node.GetVisibility() != Visible {
		return
	}
	if node.IsEnabled() && node.GetHandler() != nil {
		*result = append(*result, node)
	}
	for _, child := range node.Children() {
		collectFocusable(child, result)
	}
}

// MoveFocus 在可聚焦节点之间移动焦点。
// forward=true 时移向下一个节点（Tab），forward=false 时移向上一个节点（Shift+Tab）。
// 节点列表按深度优先遍历顺序排列，循环遍历。
func (fm *FocusManager) MoveFocus(root *Node, forward bool) {
	if root == nil {
		return
	}

	var focusable []*Node
	collectFocusable(root, &focusable)

	if len(focusable) == 0 {
		return
	}

	// 找到当前焦点节点的位置
	currentIdx := -1
	for i, n := range focusable {
		if n == fm.current {
			currentIdx = i
			break
		}
	}

	var nextIdx int
	if forward {
		// Tab：移到下一个，循环到开头
		nextIdx = (currentIdx + 1) % len(focusable)
	} else {
		// Shift+Tab：移到上一个，循环到末尾
		if currentIdx <= 0 {
			nextIdx = len(focusable) - 1
		} else {
			nextIdx = currentIdx - 1
		}
	}

	fm.RequestFocus(focusable[nextIdx])
}
