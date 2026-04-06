package editor

// EditorNode is a lightweight XML DOM node for the editor's document model.
// Unlike core.Node, it carries no runtime state (painters, handlers, layout) —
// only the tag name, XML attributes, and child structure.
type EditorNode struct {
	Tag      string            // e.g. "LinearLayout", "Button"
	Attrs    map[string]string // XML attributes in declaration order
	Children []*EditorNode
	Parent   *EditorNode

	// attrOrder preserves the original XML attribute order for round-trip fidelity.
	attrOrder []string
}

// NewEditorNode creates a node with the given tag.
func NewEditorNode(tag string) *EditorNode {
	return &EditorNode{
		Tag:   tag,
		Attrs: make(map[string]string),
	}
}

// SetAttr sets an attribute, preserving insertion order.
func (n *EditorNode) SetAttr(key, value string) {
	if _, exists := n.Attrs[key]; !exists {
		n.attrOrder = append(n.attrOrder, key)
	}
	n.Attrs[key] = value
}

// GetAttr returns an attribute value and whether it exists.
func (n *EditorNode) GetAttr(key string) (string, bool) {
	v, ok := n.Attrs[key]
	return v, ok
}

// RemoveAttr removes an attribute.
func (n *EditorNode) RemoveAttr(key string) {
	delete(n.Attrs, key)
	for i, k := range n.attrOrder {
		if k == key {
			n.attrOrder = append(n.attrOrder[:i], n.attrOrder[i+1:]...)
			break
		}
	}
}

// AttrKeys returns attribute keys in declaration order.
func (n *EditorNode) AttrKeys() []string {
	return n.attrOrder
}

// AddChild appends a child node.
func (n *EditorNode) AddChild(child *EditorNode) {
	if child.Parent != nil {
		child.Parent.RemoveChild(child)
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// InsertChild inserts a child at the given index.
func (n *EditorNode) InsertChild(index int, child *EditorNode) {
	if child.Parent != nil {
		child.Parent.RemoveChild(child)
	}
	child.Parent = n
	if index >= len(n.Children) {
		n.Children = append(n.Children, child)
		return
	}
	n.Children = append(n.Children, nil)
	copy(n.Children[index+1:], n.Children[index:])
	n.Children[index] = child
}

// RemoveChild removes a child node.
func (n *EditorNode) RemoveChild(child *EditorNode) {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}

// IndexOf returns the index of a child, or -1.
func (n *EditorNode) IndexOf(child *EditorNode) int {
	for i, c := range n.Children {
		if c == child {
			return i
		}
	}
	return -1
}

// Clone creates a deep copy of the node and its subtree.
func (n *EditorNode) Clone() *EditorNode {
	clone := &EditorNode{
		Tag:       n.Tag,
		Attrs:     make(map[string]string, len(n.Attrs)),
		attrOrder: make([]string, len(n.attrOrder)),
	}
	copy(clone.attrOrder, n.attrOrder)
	for k, v := range n.Attrs {
		clone.Attrs[k] = v
	}
	for _, child := range n.Children {
		cc := child.Clone()
		cc.Parent = clone
		clone.Children = append(clone.Children, cc)
	}
	return clone
}

// IsContainer returns true if this tag is a layout container.
func (n *EditorNode) IsContainer() bool {
	return IsContainerTag(n.Tag)
}

// DisplayName returns a short label for the node (tag + id if present).
func (n *EditorNode) DisplayName() string {
	if id, ok := n.Attrs["id"]; ok && id != "" {
		return n.Tag + " #" + id
	}
	return n.Tag
}

// Document represents an open XML layout file in the editor.
type Document struct {
	Root     *EditorNode
	FilePath string
	Modified bool
}

// NewDocument creates an empty document with a default root.
func NewDocument() *Document {
	root := NewEditorNode("LinearLayout")
	root.SetAttr("width", "match_parent")
	root.SetAttr("height", "match_parent")
	root.SetAttr("orientation", "vertical")
	return &Document{Root: root}
}
