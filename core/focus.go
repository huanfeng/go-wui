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

// ClearFocus removes focus from any node.
func (fm *FocusManager) ClearFocus() {
	fm.current = nil
}
